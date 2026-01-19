package graph

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Parser parses build files and extracts dependency information.
type Parser struct {
	graph   *Graph
	baseDir string // Base directory for path validation
}

// NewParser creates a new parser.
func NewParser() *Parser {
	return &Parser{
		graph: New(),
	}
}

// NewParserWithBaseDir creates a new parser with a base directory for path validation.
func NewParserWithBaseDir(baseDir string) *Parser {
	return &Parser{
		graph:   New(),
		baseDir: baseDir,
	}
}

// validatePath checks if the path is safe to access.
// It prevents directory traversal attacks by ensuring the path:
// 1. Is within the base directory (if set)
// 2. Has a valid build file extension
func (p *Parser) validatePath(path string) error {
	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Check file extension - only allow known build file types
	base := filepath.Base(absPath)
	validExtensions := map[string]bool{
		"Makefile":              true,
		"makefile":              true,
		"GNUmakefile":           true,
		"compile_commands.json": true,
	}
	validSuffixes := []string{".mk", ".make"}

	isValid := validExtensions[base]
	if !isValid {
		for _, suffix := range validSuffixes {
			if strings.HasSuffix(base, suffix) {
				isValid = true
				break
			}
		}
	}

	if !isValid {
		return fmt.Errorf("invalid build file: %s (must be Makefile, *.mk, or compile_commands.json)", base)
	}

	// If base directory is set, ensure path is within it
	if p.baseDir != "" {
		absBase, err := filepath.Abs(p.baseDir)
		if err != nil {
			return fmt.Errorf("invalid base directory: %w", err)
		}

		// Use filepath.Rel to check if path is within base
		rel, err := filepath.Rel(absBase, absPath)
		if err != nil {
			return fmt.Errorf("path validation failed: %w", err)
		}

		// Check for directory traversal
		if strings.HasPrefix(rel, "..") {
			return fmt.Errorf("path %s is outside allowed directory %s", path, p.baseDir)
		}
	}

	return nil
}

// ParseMakefile parses a Makefile and extracts dependencies.
func (p *Parser) ParseMakefile(path string) (*Graph, error) {
	if err := p.validatePath(path); err != nil {
		return nil, fmt.Errorf("path validation failed: %w", err)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open Makefile: %w", err)
	}
	defer file.Close()

	p.graph = New()
	scanner := bufio.NewScanner(file)

	// Regular expressions for Makefile parsing
	ruleRe := regexp.MustCompile(`^([^:=\s]+)\s*:\s*(.*)$`)
	compileRe := regexp.MustCompile(`\$\(CC\)|\$\(CXX\)|gcc|g\+\+|clang|clang\+\+`)

	var currentTarget string
	var currentDeps []string
	var inRecipe bool
	var recipeCommands []string

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			if inRecipe && currentTarget != "" {
				p.processRule(currentTarget, currentDeps, recipeCommands)
				currentTarget = ""
				currentDeps = nil
				recipeCommands = nil
				inRecipe = false
			}
			continue
		}

		// Check if this is a recipe line (starts with tab)
		if strings.HasPrefix(line, "\t") {
			inRecipe = true
			recipeCommands = append(recipeCommands, trimmed)
			continue
		}

		// If we were in a recipe and now we're not, process the rule
		if inRecipe && currentTarget != "" {
			p.processRule(currentTarget, currentDeps, recipeCommands)
			currentTarget = ""
			currentDeps = nil
			recipeCommands = nil
			inRecipe = false
		}

		// Check for rule definition
		if matches := ruleRe.FindStringSubmatch(trimmed); matches != nil {
			target := matches[1]
			deps := strings.Fields(matches[2])

			// Skip variable assignments and phony targets
			if strings.Contains(target, "=") || target == ".PHONY" {
				continue
			}

			// Check if this looks like a compile command rule
			if compileRe.MatchString(strings.Join(deps, " ")) || len(deps) > 0 {
				currentTarget = target
				currentDeps = deps
			}
		}
	}

	// Process last rule if any
	if currentTarget != "" {
		p.processRule(currentTarget, currentDeps, recipeCommands)
	}

	return p.graph, scanner.Err()
}

// processRule processes a Makefile rule and adds nodes/edges to the graph.
func (p *Parser) processRule(target string, deps []string, commands []string) {
	targetType := inferNodeType(target)
	p.graph.AddNode(&Node{
		ID:   target,
		File: target,
		Type: targetType,
	})

	for _, dep := range deps {
		// Skip variables and special targets
		if strings.HasPrefix(dep, "$") || strings.HasPrefix(dep, ".") {
			continue
		}

		depType := inferNodeType(dep)
		p.graph.AddNode(&Node{
			ID:   dep,
			File: dep,
			Type: depType,
		})

		edgeType := inferEdgeType(depType, targetType)
		p.graph.AddEdge(dep, target, edgeType)
	}
}

// ParseCompileCommands parses a compile_commands.json file.
func (p *Parser) ParseCompileCommands(path string) (*Graph, error) {
	if err := p.validatePath(path); err != nil {
		return nil, fmt.Errorf("path validation failed: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read compile_commands.json: %w", err)
	}

	var commands []CompileCommand
	if err := json.Unmarshal(data, &commands); err != nil {
		return nil, fmt.Errorf("failed to parse compile_commands.json: %w", err)
	}

	p.graph = New()

	for _, cmd := range commands {
		p.processCompileCommand(cmd)
	}

	return p.graph, nil
}

// CompileCommand represents an entry in compile_commands.json.
type CompileCommand struct {
	Directory string   `json:"directory"`
	Command   string   `json:"command"`
	File      string   `json:"file"`
	Arguments []string `json:"arguments"`
	Output    string   `json:"output"`
}

// processCompileCommand processes a compile command and extracts dependencies.
func (p *Parser) processCompileCommand(cmd CompileCommand) {
	// Get source file
	sourceFile := cmd.File
	if !filepath.IsAbs(sourceFile) {
		sourceFile = filepath.Join(cmd.Directory, sourceFile)
	}

	// Determine output file
	outputFile := cmd.Output
	if outputFile == "" {
		// Try to parse from command
		outputFile = extractOutputFile(cmd.Command)
		if outputFile == "" {
			// Default to .o extension
			base := filepath.Base(sourceFile)
			ext := filepath.Ext(base)
			outputFile = strings.TrimSuffix(base, ext) + ".o"
		}
	}

	// Extract compiler
	compiler := extractCompiler(cmd.Command, cmd.Arguments)

	// Extract include paths
	includePaths := extractIncludePaths(cmd.Command, cmd.Arguments)

	// Add source node
	p.graph.AddNode(&Node{
		ID:       sourceFile,
		File:     sourceFile,
		Type:     NodeSource,
		Compiler: compiler,
	})

	// Add object node
	p.graph.AddNode(&Node{
		ID:       outputFile,
		File:     outputFile,
		Type:     NodeObject,
		Compiler: compiler,
	})

	// Add compile edge
	p.graph.AddEdge(sourceFile, outputFile, EdgeCompilesTo)

	// Add include path nodes (simplified - would need header scanning for full accuracy)
	for _, inc := range includePaths {
		p.graph.AddNode(&Node{
			ID:   inc,
			File: inc,
			Type: NodeHeader,
		})
		p.graph.AddEdge(inc, sourceFile, EdgeIncludes)
	}
}

// inferNodeType infers the node type from the file extension.
func inferNodeType(file string) NodeType {
	ext := strings.ToLower(filepath.Ext(file))
	switch ext {
	case ".c", ".cpp", ".cc", ".cxx", ".m", ".mm":
		return NodeSource
	case ".h", ".hpp", ".hxx", ".hh":
		return NodeHeader
	case ".o", ".obj":
		return NodeObject
	case ".a", ".so", ".dylib", ".lib", ".dll":
		return NodeLibrary
	case "", ".exe", ".out":
		if !strings.Contains(file, ".") || ext == ".exe" || ext == ".out" {
			return NodeExecutable
		}
		return NodeObject
	default:
		return NodeSource
	}
}

// inferEdgeType infers the edge type from source and target node types.
func inferEdgeType(fromType, toType NodeType) EdgeType {
	if fromType == NodeHeader {
		return EdgeIncludes
	}
	if fromType == NodeSource && toType == NodeObject {
		return EdgeCompilesTo
	}
	if fromType == NodeObject && (toType == NodeExecutable || toType == NodeLibrary) {
		return EdgeLinksTo
	}
	return EdgeDependsOn
}

// extractOutputFile extracts the output file from a compile command.
func extractOutputFile(command string) string {
	// Look for -o flag
	parts := strings.Fields(command)
	for i, part := range parts {
		if part == "-o" && i+1 < len(parts) {
			return parts[i+1]
		}
		if strings.HasPrefix(part, "-o") && len(part) > 2 {
			return part[2:]
		}
	}
	return ""
}

// extractCompiler extracts the compiler from a compile command.
func extractCompiler(command string, args []string) string {
	if len(args) > 0 {
		return filepath.Base(args[0])
	}
	parts := strings.Fields(command)
	if len(parts) > 0 {
		return filepath.Base(parts[0])
	}
	return ""
}

// extractIncludePaths extracts include paths from compile arguments.
func extractIncludePaths(command string, args []string) []string {
	var paths []string
	allArgs := args
	if len(allArgs) == 0 {
		allArgs = strings.Fields(command)
	}

	for i, arg := range allArgs {
		if arg == "-I" && i+1 < len(allArgs) {
			paths = append(paths, allArgs[i+1])
		} else if strings.HasPrefix(arg, "-I") {
			paths = append(paths, arg[2:])
		}
	}
	return paths
}

// ParseAuto auto-detects the file type and parses accordingly.
func (p *Parser) ParseAuto(path string) (*Graph, error) {
	base := filepath.Base(path)

	switch {
	case base == "compile_commands.json":
		return p.ParseCompileCommands(path)
	case base == "Makefile" || strings.HasSuffix(base, ".mk"):
		return p.ParseMakefile(path)
	default:
		return nil, fmt.Errorf("unknown file type: %s", base)
	}
}
