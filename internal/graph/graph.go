package graph

import (
	"encoding/json"
	"fmt"
)

// NodeType represents the type of a build node.
type NodeType string

const (
	NodeSource     NodeType = "source"
	NodeHeader     NodeType = "header"
	NodeObject     NodeType = "object"
	NodeLibrary    NodeType = "library"
	NodeExecutable NodeType = "executable"
)

// EdgeType represents the type of dependency edge.
type EdgeType string

const (
	EdgeIncludes   EdgeType = "includes"
	EdgeCompilesTo EdgeType = "compiles_to"
	EdgeLinksTo    EdgeType = "links_to"
	EdgeDependsOn  EdgeType = "depends_on"
)

// Node represents a file or target in the build graph.
type Node struct {
	ID       string   `json:"id"`
	File     string   `json:"file"`
	Type     NodeType `json:"type"`
	Compiler string   `json:"compiler,omitempty"`
	Flags    []string `json:"flags,omitempty"`
}

// Edge represents a dependency between nodes.
type Edge struct {
	From string   `json:"from"`
	To   string   `json:"to"`
	Type EdgeType `json:"type"`
}

// Graph represents a build dependency graph.
type Graph struct {
	Nodes map[string]*Node `json:"nodes"`
	Edges []*Edge          `json:"edges"`
}

// New creates a new empty graph.
func New() *Graph {
	return &Graph{
		Nodes: make(map[string]*Node),
		Edges: make([]*Edge, 0),
	}
}

// AddNode adds a node to the graph.
func (g *Graph) AddNode(node *Node) {
	if node.ID == "" {
		node.ID = node.File
	}
	g.Nodes[node.ID] = node
}

// AddEdge adds an edge to the graph.
func (g *Graph) AddEdge(from, to string, edgeType EdgeType) {
	g.Edges = append(g.Edges, &Edge{
		From: from,
		To:   to,
		Type: edgeType,
	})
}

// GetNode returns a node by ID.
func (g *Graph) GetNode(id string) *Node {
	return g.Nodes[id]
}

// NodeCount returns the number of nodes.
func (g *Graph) NodeCount() int {
	return len(g.Nodes)
}

// EdgeCount returns the number of edges.
func (g *Graph) EdgeCount() int {
	return len(g.Edges)
}

// GetNodesByType returns all nodes of a given type.
func (g *Graph) GetNodesByType(nodeType NodeType) []*Node {
	var nodes []*Node
	for _, node := range g.Nodes {
		if node.Type == nodeType {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// GetIncomingEdges returns all edges pointing to a node.
func (g *Graph) GetIncomingEdges(nodeID string) []*Edge {
	var edges []*Edge
	for _, edge := range g.Edges {
		if edge.To == nodeID {
			edges = append(edges, edge)
		}
	}
	return edges
}

// GetOutgoingEdges returns all edges from a node.
func (g *Graph) GetOutgoingEdges(nodeID string) []*Edge {
	var edges []*Edge
	for _, edge := range g.Edges {
		if edge.From == nodeID {
			edges = append(edges, edge)
		}
	}
	return edges
}

// ToJSON converts the graph to JSON.
func (g *Graph) ToJSON() ([]byte, error) {
	return json.MarshalIndent(g, "", "  ")
}

// ToDOT converts the graph to DOT format for Graphviz.
func (g *Graph) ToDOT() string {
	dot := "digraph BuildGraph {\n"
	dot += "  rankdir=LR;\n"
	dot += "  node [shape=box];\n\n"

	// Define node styles by type
	dot += "  // Node styles\n"
	dot += "  node [style=filled];\n"

	// Add nodes with colors by type
	for _, node := range g.Nodes {
		color := getNodeColor(node.Type)
		shape := getNodeShape(node.Type)
		label := node.File
		if len(label) > 30 {
			label = "..." + label[len(label)-27:]
		}
		dot += fmt.Sprintf("  \"%s\" [label=\"%s\", fillcolor=\"%s\", shape=%s];\n",
			node.ID, label, color, shape)
	}

	dot += "\n  // Edges\n"

	// Add edges with styles by type
	for _, edge := range g.Edges {
		style := getEdgeStyle(edge.Type)
		dot += fmt.Sprintf("  \"%s\" -> \"%s\" [%s];\n", edge.From, edge.To, style)
	}

	dot += "}\n"
	return dot
}

// Merge merges another graph into this one.
func (g *Graph) Merge(other *Graph) {
	for id, node := range other.Nodes {
		if _, exists := g.Nodes[id]; !exists {
			g.Nodes[id] = node
		}
	}
	g.Edges = append(g.Edges, other.Edges...)
}

// getNodeColor returns the color for a node type.
func getNodeColor(nodeType NodeType) string {
	switch nodeType {
	case NodeSource:
		return "#a8d8ea"
	case NodeHeader:
		return "#d4f1f4"
	case NodeObject:
		return "#ffcc80"
	case NodeLibrary:
		return "#c5e1a5"
	case NodeExecutable:
		return "#ef9a9a"
	default:
		return "#ffffff"
	}
}

// getNodeShape returns the shape for a node type.
func getNodeShape(nodeType NodeType) string {
	switch nodeType {
	case NodeSource:
		return "note"
	case NodeHeader:
		return "tab"
	case NodeObject:
		return "box"
	case NodeLibrary:
		return "folder"
	case NodeExecutable:
		return "octagon"
	default:
		return "box"
	}
}

// getEdgeStyle returns the DOT style for an edge type.
func getEdgeStyle(edgeType EdgeType) string {
	switch edgeType {
	case EdgeIncludes:
		return "style=dashed, color=gray"
	case EdgeCompilesTo:
		return "style=solid, color=blue"
	case EdgeLinksTo:
		return "style=bold, color=green"
	case EdgeDependsOn:
		return "style=dotted, color=orange"
	default:
		return ""
	}
}
