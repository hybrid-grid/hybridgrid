package graph

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	g := New()
	if g == nil {
		t.Fatal("New() returned nil")
	}
	if g.Nodes == nil {
		t.Error("Nodes map should not be nil")
	}
	if g.Edges == nil {
		t.Error("Edges slice should not be nil")
	}
	if g.NodeCount() != 0 {
		t.Errorf("Expected 0 nodes, got %d", g.NodeCount())
	}
	if g.EdgeCount() != 0 {
		t.Errorf("Expected 0 edges, got %d", g.EdgeCount())
	}
}

func TestAddNode(t *testing.T) {
	g := New()

	node := &Node{
		ID:       "test.c",
		File:     "test.c",
		Type:     NodeSource,
		Compiler: "gcc",
	}
	g.AddNode(node)

	if g.NodeCount() != 1 {
		t.Errorf("Expected 1 node, got %d", g.NodeCount())
	}

	retrieved := g.GetNode("test.c")
	if retrieved == nil {
		t.Fatal("GetNode returned nil")
	}
	if retrieved.File != "test.c" {
		t.Errorf("Expected File=test.c, got %s", retrieved.File)
	}
	if retrieved.Type != NodeSource {
		t.Errorf("Expected Type=source, got %s", retrieved.Type)
	}
}

func TestAddNodeAutoID(t *testing.T) {
	g := New()

	// Node with empty ID should use File as ID
	node := &Node{
		File: "auto.c",
		Type: NodeSource,
	}
	g.AddNode(node)

	retrieved := g.GetNode("auto.c")
	if retrieved == nil {
		t.Fatal("Node with auto-generated ID not found")
	}
}

func TestAddEdge(t *testing.T) {
	g := New()

	g.AddNode(&Node{ID: "main.c", File: "main.c", Type: NodeSource})
	g.AddNode(&Node{ID: "main.o", File: "main.o", Type: NodeObject})
	g.AddEdge("main.c", "main.o", EdgeCompilesTo)

	if g.EdgeCount() != 1 {
		t.Errorf("Expected 1 edge, got %d", g.EdgeCount())
	}

	edge := g.Edges[0]
	if edge.From != "main.c" {
		t.Errorf("Expected From=main.c, got %s", edge.From)
	}
	if edge.To != "main.o" {
		t.Errorf("Expected To=main.o, got %s", edge.To)
	}
	if edge.Type != EdgeCompilesTo {
		t.Errorf("Expected Type=compiles_to, got %s", edge.Type)
	}
}

func TestGetNodesByType(t *testing.T) {
	g := New()

	g.AddNode(&Node{ID: "a.c", File: "a.c", Type: NodeSource})
	g.AddNode(&Node{ID: "b.c", File: "b.c", Type: NodeSource})
	g.AddNode(&Node{ID: "a.o", File: "a.o", Type: NodeObject})
	g.AddNode(&Node{ID: "main", File: "main", Type: NodeExecutable})

	sources := g.GetNodesByType(NodeSource)
	if len(sources) != 2 {
		t.Errorf("Expected 2 source nodes, got %d", len(sources))
	}

	objects := g.GetNodesByType(NodeObject)
	if len(objects) != 1 {
		t.Errorf("Expected 1 object node, got %d", len(objects))
	}

	executables := g.GetNodesByType(NodeExecutable)
	if len(executables) != 1 {
		t.Errorf("Expected 1 executable node, got %d", len(executables))
	}

	headers := g.GetNodesByType(NodeHeader)
	if len(headers) != 0 {
		t.Errorf("Expected 0 header nodes, got %d", len(headers))
	}
}

func TestGetIncomingEdges(t *testing.T) {
	g := New()

	g.AddNode(&Node{ID: "a.c", File: "a.c", Type: NodeSource})
	g.AddNode(&Node{ID: "b.c", File: "b.c", Type: NodeSource})
	g.AddNode(&Node{ID: "main.o", File: "main.o", Type: NodeObject})

	g.AddEdge("a.c", "main.o", EdgeCompilesTo)
	g.AddEdge("b.c", "main.o", EdgeCompilesTo)

	incoming := g.GetIncomingEdges("main.o")
	if len(incoming) != 2 {
		t.Errorf("Expected 2 incoming edges, got %d", len(incoming))
	}

	incoming = g.GetIncomingEdges("a.c")
	if len(incoming) != 0 {
		t.Errorf("Expected 0 incoming edges for a.c, got %d", len(incoming))
	}
}

func TestGetOutgoingEdges(t *testing.T) {
	g := New()

	g.AddNode(&Node{ID: "main.c", File: "main.c", Type: NodeSource})
	g.AddNode(&Node{ID: "main.o", File: "main.o", Type: NodeObject})
	g.AddNode(&Node{ID: "main", File: "main", Type: NodeExecutable})

	g.AddEdge("main.c", "main.o", EdgeCompilesTo)
	g.AddEdge("main.o", "main", EdgeLinksTo)

	outgoing := g.GetOutgoingEdges("main.c")
	if len(outgoing) != 1 {
		t.Errorf("Expected 1 outgoing edge from main.c, got %d", len(outgoing))
	}

	outgoing = g.GetOutgoingEdges("main.o")
	if len(outgoing) != 1 {
		t.Errorf("Expected 1 outgoing edge from main.o, got %d", len(outgoing))
	}

	outgoing = g.GetOutgoingEdges("main")
	if len(outgoing) != 0 {
		t.Errorf("Expected 0 outgoing edges from main, got %d", len(outgoing))
	}
}

func TestToJSON(t *testing.T) {
	g := New()
	g.AddNode(&Node{ID: "test.c", File: "test.c", Type: NodeSource})
	g.AddEdge("test.c", "test.o", EdgeCompilesTo)

	jsonData, err := g.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// Verify it's valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	if _, ok := result["nodes"]; !ok {
		t.Error("JSON missing 'nodes' field")
	}
	if _, ok := result["edges"]; !ok {
		t.Error("JSON missing 'edges' field")
	}
}

func TestToDOT(t *testing.T) {
	g := New()
	g.AddNode(&Node{ID: "main.c", File: "main.c", Type: NodeSource})
	g.AddNode(&Node{ID: "main.o", File: "main.o", Type: NodeObject})
	g.AddEdge("main.c", "main.o", EdgeCompilesTo)

	dot := g.ToDOT()

	if !strings.Contains(dot, "digraph BuildGraph") {
		t.Error("DOT output missing digraph declaration")
	}
	if !strings.Contains(dot, "main.c") {
		t.Error("DOT output missing main.c node")
	}
	if !strings.Contains(dot, "main.o") {
		t.Error("DOT output missing main.o node")
	}
	if !strings.Contains(dot, "->") {
		t.Error("DOT output missing edge")
	}
}

func TestMerge(t *testing.T) {
	g1 := New()
	g1.AddNode(&Node{ID: "a.c", File: "a.c", Type: NodeSource})
	g1.AddEdge("a.c", "a.o", EdgeCompilesTo)

	g2 := New()
	g2.AddNode(&Node{ID: "b.c", File: "b.c", Type: NodeSource})
	g2.AddEdge("b.c", "b.o", EdgeCompilesTo)

	g1.Merge(g2)

	if g1.NodeCount() != 2 {
		t.Errorf("Expected 2 nodes after merge, got %d", g1.NodeCount())
	}
	if g1.EdgeCount() != 2 {
		t.Errorf("Expected 2 edges after merge, got %d", g1.EdgeCount())
	}
}

func TestMergeSkipsDuplicateNodes(t *testing.T) {
	g1 := New()
	g1.AddNode(&Node{ID: "a.c", File: "a.c", Type: NodeSource})

	g2 := New()
	g2.AddNode(&Node{ID: "a.c", File: "a.c", Type: NodeSource}) // Duplicate

	g1.Merge(g2)

	if g1.NodeCount() != 1 {
		t.Errorf("Expected 1 node after merge (no duplicates), got %d", g1.NodeCount())
	}
}

func TestNodeTypes(t *testing.T) {
	tests := []struct {
		nodeType NodeType
		expected string
	}{
		{NodeSource, "source"},
		{NodeHeader, "header"},
		{NodeObject, "object"},
		{NodeLibrary, "library"},
		{NodeExecutable, "executable"},
	}

	for _, tt := range tests {
		if string(tt.nodeType) != tt.expected {
			t.Errorf("NodeType %v != %s", tt.nodeType, tt.expected)
		}
	}
}

func TestEdgeTypes(t *testing.T) {
	tests := []struct {
		edgeType EdgeType
		expected string
	}{
		{EdgeIncludes, "includes"},
		{EdgeCompilesTo, "compiles_to"},
		{EdgeLinksTo, "links_to"},
		{EdgeDependsOn, "depends_on"},
	}

	for _, tt := range tests {
		if string(tt.edgeType) != tt.expected {
			t.Errorf("EdgeType %v != %s", tt.edgeType, tt.expected)
		}
	}
}

func TestGetNodeColor(t *testing.T) {
	tests := []struct {
		nodeType NodeType
		wantLen  int // Just verify it returns a color string
	}{
		{NodeSource, 7},
		{NodeHeader, 7},
		{NodeObject, 7},
		{NodeLibrary, 7},
		{NodeExecutable, 7},
	}

	for _, tt := range tests {
		color := getNodeColor(tt.nodeType)
		if len(color) != tt.wantLen {
			t.Errorf("getNodeColor(%s) = %s, want length %d", tt.nodeType, color, tt.wantLen)
		}
	}
}

func TestGetNodeShape(t *testing.T) {
	tests := []struct {
		nodeType NodeType
		want     string
	}{
		{NodeSource, "note"},
		{NodeHeader, "tab"},
		{NodeObject, "box"},
		{NodeLibrary, "folder"},
		{NodeExecutable, "octagon"},
	}

	for _, tt := range tests {
		got := getNodeShape(tt.nodeType)
		if got != tt.want {
			t.Errorf("getNodeShape(%s) = %s, want %s", tt.nodeType, got, tt.want)
		}
	}
}

func TestGetEdgeStyle(t *testing.T) {
	tests := []struct {
		edgeType EdgeType
		wantLen  int // Just verify it returns a style string
	}{
		{EdgeIncludes, 1},
		{EdgeCompilesTo, 1},
		{EdgeLinksTo, 1},
		{EdgeDependsOn, 1},
	}

	for _, tt := range tests {
		style := getEdgeStyle(tt.edgeType)
		if len(style) < tt.wantLen {
			t.Errorf("getEdgeStyle(%s) = %s, want length >= %d", tt.edgeType, style, tt.wantLen)
		}
	}
}

func TestToDOTLongLabel(t *testing.T) {
	g := New()
	// Add node with very long filename
	longName := "very_long_filename_that_exceeds_thirty_characters.c"
	g.AddNode(&Node{ID: longName, File: longName, Type: NodeSource})

	dot := g.ToDOT()

	// Should truncate label
	if !strings.Contains(dot, "...") {
		t.Error("Long label should be truncated with ...")
	}
}
