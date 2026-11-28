// Package tscbuild runs a TypeScript project-references build
// (`tsc -b`) across Hybrid-Grid workers: every project in the reference
// graph becomes one `tsc -p` task, sent as soon as the projects it
// references have been built, so independent projects compile in
// parallel on different machines.
package tscbuild

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Project is one tsconfig in the reference graph.
type Project struct {
	// Config is the tsconfig path relative to the graph root, slash-separated.
	Config string
	// Dir is Config's directory relative to the root ("." for the root).
	Dir string
	// Refs are the Config values of the projects this one references.
	Refs []string
}

// Graph is a project-references build rooted at one directory.
type Graph struct {
	// Root is the absolute directory of the root tsconfig; every project
	// must live under it, and it is the root of each task's archive.
	Root string
	// Projects to build, in a valid build order (references first).
	Projects []*Project
}

type tsconfig struct {
	Files      *[]string `json:"files"`
	Include    []string  `json:"include"`
	References []struct {
		Path string `json:"path"`
	} `json:"references"`
}

// Load reads the tsconfig at path (a file, or a directory containing
// tsconfig.json) and follows its references. A root that only lists
// references ("files": [] and no "include") is a solution file and is
// not built itself.
//
// Graph.Root is the workspace root: the nearest directory at or above
// the tsconfig holding a lockfile, .git or node_modules, so a project
// can be built on its own even when it references siblings
// (packages/app -> ../lib) or extends a shared config further up.
func Load(path string) (*Graph, error) {
	rootCfg, err := configPath(path)
	if err != nil {
		return nil, err
	}
	type node struct {
		abs  string
		refs []string // absolute tsconfig paths
	}
	var (
		order    []*node // post-order: references first
		seen     = make(map[string]bool)
		visiting = make(map[string]bool)
	)
	var visit func(abs string, isRoot bool) error
	visit = func(abs string, isRoot bool) error {
		if seen[abs] {
			return nil
		}
		if visiting[abs] {
			return fmt.Errorf("project references form a cycle through %s", abs)
		}
		visiting[abs] = true
		defer delete(visiting, abs)

		cfg, err := readTsconfig(abs)
		if err != nil {
			return err
		}
		n := &node{abs: abs}
		for _, ref := range cfg.References {
			if strings.TrimSpace(ref.Path) == "" {
				continue
			}
			refAbs, err := configPath(filepath.Join(filepath.Dir(abs), filepath.FromSlash(ref.Path)))
			if err != nil {
				return fmt.Errorf("%s: %w", abs, err)
			}
			if err := visit(refAbs, false); err != nil {
				return err
			}
			n.refs = append(n.refs, refAbs)
		}
		seen[abs] = true
		if isRoot && cfg.Files != nil && len(*cfg.Files) == 0 && cfg.Include == nil {
			return nil // a solution file: only its references are built
		}
		order = append(order, n)
		return nil
	}
	if err := visit(rootCfg, true); err != nil {
		return nil, err
	}
	if len(order) == 0 {
		return nil, errors.New("no TypeScript projects to build")
	}

	dirs := make([]string, 0, len(order)+1)
	dirs = append(dirs, filepath.Dir(rootCfg))
	for _, n := range order {
		dirs = append(dirs, filepath.Dir(n.abs))
	}
	g := &Graph{Root: workspaceRoot(filepath.Dir(rootCfg), dirs)}
	rel := func(abs string) string {
		r, _ := filepath.Rel(g.Root, abs) // Root contains every project
		return filepath.ToSlash(r)
	}
	for _, n := range order {
		cfg := rel(n.abs)
		p := &Project{Config: cfg, Dir: filepath.ToSlash(filepath.Dir(filepath.FromSlash(cfg)))}
		for _, r := range n.refs {
			p.Refs = append(p.Refs, rel(r))
		}
		g.Projects = append(g.Projects, p)
	}
	return g, nil
}

// workspaceMarkers identify the root of a JavaScript workspace.
var workspaceMarkers = []string{
	"package-lock.json", "pnpm-lock.yaml", "pnpm-workspace.yaml", "yarn.lock", "bun.lockb", "bun.lock",
	".git", "node_modules",
}

// workspaceRoot walks up from start to the nearest directory holding a
// workspace marker, then widens it until it contains every dir.
func workspaceRoot(start string, dirs []string) string {
	root := ""
	for d := start; ; d = filepath.Dir(d) {
		for _, m := range workspaceMarkers {
			if _, err := os.Stat(filepath.Join(d, m)); err == nil {
				root = d
				break
			}
		}
		if root != "" || filepath.Dir(d) == d {
			break
		}
	}
	if root == "" {
		root = start
	}
	for _, d := range dirs {
		for !contains(root, d) {
			root = filepath.Dir(root)
		}
	}
	return root
}

func contains(base, path string) bool {
	rel, err := filepath.Rel(base, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// Closure returns the Config values of p and everything it references,
// directly or not.
func (g *Graph) Closure(p *Project) map[string]bool {
	byConfig := make(map[string]*Project, len(g.Projects))
	for _, q := range g.Projects {
		byConfig[q.Config] = q
	}
	out := make(map[string]bool)
	var walk func(cfg string)
	walk = func(cfg string) {
		if out[cfg] {
			return
		}
		out[cfg] = true
		if q := byConfig[cfg]; q != nil {
			for _, r := range q.Refs {
				walk(r)
			}
		}
	}
	walk(p.Config)
	return out
}

// configPath resolves a tsconfig file or a directory holding tsconfig.json.
func configPath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("tsconfig not found: %w", err)
	}
	if info.IsDir() {
		abs = filepath.Join(abs, "tsconfig.json")
		if _, err := os.Stat(abs); err != nil {
			return "", fmt.Errorf("tsconfig not found: %w", err)
		}
	}
	return abs, nil
}

func readTsconfig(path string) (*tsconfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}
	var cfg tsconfig
	if err := json.Unmarshal(stripJSONC(data), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}
	return &cfg, nil
}

// stripJSONC turns tsconfig's JSON-with-comments into JSON: it drops
// // and /* */ comments outside strings and trailing commas.
func stripJSONC(in []byte) []byte {
	var out bytes.Buffer
	inString, escaped := false, false
	for i := 0; i < len(in); i++ {
		c := in[i]
		if inString {
			out.WriteByte(c)
			switch {
			case escaped:
				escaped = false
			case c == '\\':
				escaped = true
			case c == '"':
				inString = false
			}
			continue
		}
		switch {
		case c == '"':
			inString = true
			out.WriteByte(c)
		case c == '/' && i+1 < len(in) && in[i+1] == '/':
			for i < len(in) && in[i] != '\n' {
				i++
			}
			out.WriteByte('\n')
		case c == '/' && i+1 < len(in) && in[i+1] == '*':
			i += 2
			for i+1 < len(in) && !(in[i] == '*' && in[i+1] == '/') {
				i++
			}
			i++
		case c == ',':
			j := i + 1
			for j < len(in) && strings.ContainsRune(" \t\r\n", rune(in[j])) {
				j++
			}
			if j < len(in) && (in[j] == '}' || in[j] == ']') {
				continue // trailing comma
			}
			out.WriteByte(c)
		default:
			out.WriteByte(c)
		}
	}
	return out.Bytes()
}

// projectDirs maps every project directory to whether any project in
// that directory is in closure.
func (g *Graph) projectDirs(closure map[string]bool) map[string]bool {
	dirs := make(map[string]bool)
	for _, p := range g.Projects {
		dirs[p.Dir] = dirs[p.Dir] || closure[p.Config]
	}
	return dirs
}
