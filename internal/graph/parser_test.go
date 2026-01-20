package graph

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewParser(t *testing.T) {
	p := NewParser()
	if p == nil {
		t.Fatal("NewParser returned nil")
	}
	if p.graph == nil {
		t.Error("Parser graph should not be nil")
	}
}

func TestNewParserWithBaseDir(t *testing.T) {
	p := NewParserWithBaseDir("/tmp")
	if p == nil {
		t.Fatal("NewParserWithBaseDir returned nil")
	}
	if p.baseDir != "/tmp" {
		t.Errorf("Expected baseDir=/tmp, got %s", p.baseDir)
	}
}

func TestValidatePath(t *testing.T) {
	p := NewParser()

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid Makefile", "Makefile", false},
		{"valid makefile", "makefile", false},
		{"valid GNUmakefile", "GNUmakefile", false},
		{"valid .mk file", "build.mk", false},
		{"valid compile_commands.json", "compile_commands.json", false},
		{"invalid txt file", "file.txt", true},
		{"invalid c file", "main.c", true},
		{"invalid passwd", "passwd", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, tt.path)
			if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			err := p.validatePath(tmpFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePath(%s) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePathWithBaseDir(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewParserWithBaseDir(tmpDir)

	// Create valid file inside base dir
	validPath := filepath.Join(tmpDir, "Makefile")
	if err := os.WriteFile(validPath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Should succeed - file is inside base dir
	if err := p.validatePath(validPath); err != nil {
		t.Errorf("validatePath should succeed for file inside base dir: %v", err)
	}
}

func TestInferNodeType(t *testing.T) {
	tests := []struct {
		file     string
		expected NodeType
	}{
		{"main.c", NodeSource},
		{"main.cpp", NodeSource},
		{"main.cc", NodeSource},
		{"main.cxx", NodeSource},
		{"main.m", NodeSource},
		{"main.mm", NodeSource},
		{"header.h", NodeHeader},
		{"header.hpp", NodeHeader},
		{"header.hxx", NodeHeader},
		{"header.hh", NodeHeader},
		{"main.o", NodeObject},
		{"main.obj", NodeObject},
		{"libfoo.a", NodeLibrary},
		{"libfoo.so", NodeLibrary},
		{"libfoo.dylib", NodeLibrary},
		{"foo.lib", NodeLibrary},
		{"foo.dll", NodeLibrary},
		{"main", NodeExecutable},
		{"main.exe", NodeExecutable},
		{"main.out", NodeExecutable},
		{"unknown.xyz", NodeSource}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			got := inferNodeType(tt.file)
			if got != tt.expected {
				t.Errorf("inferNodeType(%s) = %s, want %s", tt.file, got, tt.expected)
			}
		})
	}
}

func TestInferEdgeType(t *testing.T) {
	tests := []struct {
		from     NodeType
		to       NodeType
		expected EdgeType
	}{
		{NodeHeader, NodeSource, EdgeIncludes},
		{NodeSource, NodeObject, EdgeCompilesTo},
		{NodeObject, NodeExecutable, EdgeLinksTo},
		{NodeObject, NodeLibrary, EdgeLinksTo},
		{NodeSource, NodeExecutable, EdgeDependsOn}, // Default
	}

	for _, tt := range tests {
		t.Run(string(tt.from)+"->"+string(tt.to), func(t *testing.T) {
			got := inferEdgeType(tt.from, tt.to)
			if got != tt.expected {
				t.Errorf("inferEdgeType(%s, %s) = %s, want %s", tt.from, tt.to, got, tt.expected)
			}
		})
	}
}

func TestExtractOutputFile(t *testing.T) {
	tests := []struct {
		command  string
		expected string
	}{
		{"gcc -c main.c -o main.o", "main.o"},
		{"gcc -c main.c -omain.o", "main.o"},
		{"gcc -c main.c", ""},
		{"clang++ -std=c++17 -o app main.cpp", "app"},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			got := extractOutputFile(tt.command)
			if got != tt.expected {
				t.Errorf("extractOutputFile(%s) = %s, want %s", tt.command, got, tt.expected)
			}
		})
	}
}

func TestExtractCompiler(t *testing.T) {
	tests := []struct {
		command  string
		args     []string
		expected string
	}{
		{"gcc -c main.c", nil, "gcc"},
		{"/usr/bin/clang++ -c main.cpp", nil, "clang++"},
		{"", []string{"/usr/bin/gcc", "-c", "main.c"}, "gcc"},
		{"", []string{"clang", "-c", "main.c"}, "clang"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := extractCompiler(tt.command, tt.args)
			if got != tt.expected {
				t.Errorf("extractCompiler(%s, %v) = %s, want %s", tt.command, tt.args, got, tt.expected)
			}
		})
	}
}

func TestExtractIncludePaths(t *testing.T) {
	tests := []struct {
		command  string
		args     []string
		expected []string
	}{
		{"gcc -I/usr/include -c main.c", nil, []string{"/usr/include"}},
		{"gcc -I /usr/local/include -c main.c", nil, []string{"/usr/local/include"}},
		{"gcc -I./include -I../lib -c main.c", nil, []string{"./include", "../lib"}},
		{"", []string{"gcc", "-I", "/include", "-c", "main.c"}, []string{"/include"}},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			got := extractIncludePaths(tt.command, tt.args)
			if len(got) != len(tt.expected) {
				t.Errorf("extractIncludePaths() = %v, want %v", got, tt.expected)
				return
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("extractIncludePaths()[%d] = %s, want %s", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestParseMakefile(t *testing.T) {
	tmpDir := t.TempDir()
	makefile := filepath.Join(tmpDir, "Makefile")

	content := `
all: main

main: main.o utils.o
	gcc -o main main.o utils.o

main.o: main.c
	gcc -c main.c

utils.o: utils.c
	gcc -c utils.c

.PHONY: clean
clean:
	rm -f *.o main
`
	if err := os.WriteFile(makefile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create Makefile: %v", err)
	}

	p := NewParser()
	g, err := p.ParseMakefile(makefile)
	if err != nil {
		t.Fatalf("ParseMakefile failed: %v", err)
	}

	if g.NodeCount() == 0 {
		t.Error("Expected nodes from Makefile")
	}
}

func TestParseMakefileInvalid(t *testing.T) {
	p := NewParser()
	_, err := p.ParseMakefile("/nonexistent/Makefile")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestParseCompileCommands(t *testing.T) {
	tmpDir := t.TempDir()
	ccFile := filepath.Join(tmpDir, "compile_commands.json")

	content := `[
		{
			"directory": "/tmp/build",
			"command": "gcc -c -o main.o main.c",
			"file": "main.c"
		},
		{
			"directory": "/tmp/build",
			"command": "gcc -c -o utils.o utils.c",
			"file": "utils.c"
		}
	]`
	if err := os.WriteFile(ccFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create compile_commands.json: %v", err)
	}

	p := NewParser()
	g, err := p.ParseCompileCommands(ccFile)
	if err != nil {
		t.Fatalf("ParseCompileCommands failed: %v", err)
	}

	// Should have source and object nodes
	if g.NodeCount() < 2 {
		t.Errorf("Expected at least 2 nodes, got %d", g.NodeCount())
	}
}

func TestParseCompileCommandsInvalid(t *testing.T) {
	p := NewParser()
	_, err := p.ParseCompileCommands("/nonexistent/compile_commands.json")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestParseCompileCommandsInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	ccFile := filepath.Join(tmpDir, "compile_commands.json")

	if err := os.WriteFile(ccFile, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	p := NewParser()
	_, err := p.ParseCompileCommands(ccFile)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestParseAuto(t *testing.T) {
	tmpDir := t.TempDir()

	// Test Makefile detection
	makefile := filepath.Join(tmpDir, "Makefile")
	if err := os.WriteFile(makefile, []byte("all:\n\techo hello"), 0644); err != nil {
		t.Fatalf("Failed to create Makefile: %v", err)
	}

	p := NewParser()
	g, err := p.ParseAuto(makefile)
	if err != nil {
		t.Errorf("ParseAuto(Makefile) failed: %v", err)
	}
	if g == nil {
		t.Error("ParseAuto(Makefile) returned nil graph")
	}

	// Test compile_commands.json detection
	ccFile := filepath.Join(tmpDir, "compile_commands.json")
	if err := os.WriteFile(ccFile, []byte("[]"), 0644); err != nil {
		t.Fatalf("Failed to create compile_commands.json: %v", err)
	}

	g, err = p.ParseAuto(ccFile)
	if err != nil {
		t.Errorf("ParseAuto(compile_commands.json) failed: %v", err)
	}
	if g == nil {
		t.Error("ParseAuto(compile_commands.json) returned nil graph")
	}

	// Test unknown file type
	unknownFile := filepath.Join(tmpDir, "unknown.txt")
	if err := os.WriteFile(unknownFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create unknown.txt: %v", err)
	}

	_, err = p.ParseAuto(unknownFile)
	if err == nil {
		t.Error("ParseAuto should fail for unknown file type")
	}
}

func TestProcessCompileCommand(t *testing.T) {
	p := NewParser()
	p.graph = New()

	cmd := CompileCommand{
		Directory: "/tmp/build",
		Command:   "gcc -I./include -c -o main.o main.c",
		File:      "main.c",
		Output:    "main.o",
	}

	p.processCompileCommand(cmd)

	// Should have source and object nodes
	if p.graph.NodeCount() < 2 {
		t.Errorf("Expected at least 2 nodes, got %d", p.graph.NodeCount())
	}

	// Should have at least one edge (source -> object)
	if p.graph.EdgeCount() < 1 {
		t.Errorf("Expected at least 1 edge, got %d", p.graph.EdgeCount())
	}
}

func TestProcessCompileCommandNoOutput(t *testing.T) {
	p := NewParser()
	p.graph = New()

	cmd := CompileCommand{
		Directory: "/tmp/build",
		Command:   "gcc -c main.c",
		File:      "main.c",
		// No Output specified - should be inferred
	}

	p.processCompileCommand(cmd)

	// Should have source and object nodes
	if p.graph.NodeCount() < 2 {
		t.Errorf("Expected at least 2 nodes, got %d", p.graph.NodeCount())
	}
}

func TestProcessCompileCommandWithArguments(t *testing.T) {
	p := NewParser()
	p.graph = New()

	cmd := CompileCommand{
		Directory: "/tmp/build",
		File:      "main.c",
		Arguments: []string{"gcc", "-c", "-o", "main.o", "main.c"},
	}

	p.processCompileCommand(cmd)

	// Should have nodes
	if p.graph.NodeCount() < 2 {
		t.Errorf("Expected at least 2 nodes, got %d", p.graph.NodeCount())
	}
}

func TestParseMakefileWithLineContinuation(t *testing.T) {
	tmpDir := t.TempDir()
	makefile := filepath.Join(tmpDir, "Makefile")

	// Makefile with line continuations
	content := `SRCS = main.c \
       utils.c \
       helper.c

OBJS = $(SRCS:.c=.o)

all: main

main: main.o utils.o helper.o
	gcc -o main main.o utils.o helper.o

main.o: main.c
	gcc -c main.c

utils.o: utils.c
	gcc -c utils.c
`
	if err := os.WriteFile(makefile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create Makefile: %v", err)
	}

	p := NewParser()
	g, err := p.ParseMakefile(makefile)
	if err != nil {
		t.Fatalf("ParseMakefile failed: %v", err)
	}

	// Should have nodes from the parsed rules
	if g.NodeCount() == 0 {
		t.Error("Expected nodes from Makefile with line continuations")
	}

	// Verify main target exists
	mainNode := g.GetNode("main")
	if mainNode == nil {
		t.Error("Expected 'main' node")
	}
}

func TestParseMakefileWithPatternRule(t *testing.T) {
	tmpDir := t.TempDir()
	makefile := filepath.Join(tmpDir, "Makefile")

	// Makefile with pattern rule
	content := `CC = gcc
CFLAGS = -Wall -O2

SRCS = main.c utils.c
OBJS = main.o utils.o

all: main

main: $(OBJS)
	$(CC) -o main $(OBJS)

%.o: %.c
	$(CC) $(CFLAGS) -c $< -o $@

.PHONY: clean
clean:
	rm -f *.o main
`
	if err := os.WriteFile(makefile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create Makefile: %v", err)
	}

	p := NewParser()
	g, err := p.ParseMakefile(makefile)
	if err != nil {
		t.Fatalf("ParseMakefile failed: %v", err)
	}

	// Should have pattern rule node
	patternNode := g.GetNode("%.o")
	if patternNode == nil {
		t.Error("Expected pattern rule node for object files")
	}

	// Pattern node should be object type
	if patternNode != nil && patternNode.Type != NodeObject {
		t.Errorf("Expected pattern node type 'object', got '%s'", patternNode.Type)
	}

	// Should have %.c source node
	sourcePattern := g.GetNode("%.c")
	if sourcePattern == nil {
		t.Error("Expected pattern source node for C files")
	}
}

func TestInferPatternNodeType(t *testing.T) {
	tests := []struct {
		pattern  string
		expected NodeType
	}{
		{"%.o", NodeObject},
		{"%.c", NodeSource},
		{"%.cpp", NodeSource},
		{"%.h", NodeHeader},
		{"%.a", NodeLibrary},
		{"%", NodeExecutable},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			got := inferPatternNodeType(tt.pattern)
			if got != tt.expected {
				t.Errorf("inferPatternNodeType(%s) = %s, want %s", tt.pattern, got, tt.expected)
			}
		})
	}
}

func TestProcessPatternRule(t *testing.T) {
	p := NewParser()
	p.graph = New()

	p.processPatternRule("%.o", "%.c", nil, []string{"$(CC) -c $< -o $@"})

	// Should have pattern nodes
	if p.graph.NodeCount() < 2 {
		t.Errorf("Expected at least 2 nodes, got %d", p.graph.NodeCount())
	}

	// Should have edge from %.c to %.o
	if p.graph.EdgeCount() < 1 {
		t.Errorf("Expected at least 1 edge, got %d", p.graph.EdgeCount())
	}

	// Verify edge type is compiles_to
	edges := p.graph.GetIncomingEdges("%.o")
	if len(edges) == 0 {
		t.Error("Expected incoming edge for pattern object node")
	} else if edges[0].Type != EdgeCompilesTo {
		t.Errorf("Expected edge type 'compiles_to', got '%s'", edges[0].Type)
	}
}

func TestReadLinesWithContinuation(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")

	content := `line1
line2 \
continued
line3
multi \
line \
continuation
end
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	file, err := os.Open(tmpFile)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	p := NewParser()
	lines, err := p.readLinesWithContinuation(file)
	if err != nil {
		t.Fatalf("Failed to read lines: %v", err)
	}

	// Verify line continuations are handled
	if len(lines) != 5 {
		t.Errorf("Expected 5 lines, got %d", len(lines))
	}

	// Verify first line is unchanged
	if lines[0] != "line1" {
		t.Errorf("Expected 'line1', got '%s'", lines[0])
	}

	// Verify continuation is joined
	if !strings.Contains(lines[1], "line2") || !strings.Contains(lines[1], "continued") {
		t.Errorf("Expected joined continuation line, got '%s'", lines[1])
	}

	// Verify multi-line continuation
	if !strings.Contains(lines[3], "multi") && !strings.Contains(lines[3], "continuation") {
		t.Errorf("Expected multi-line continuation, got '%s'", lines[3])
	}
}
