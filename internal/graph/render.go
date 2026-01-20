package graph

import (
	"encoding/json"
	"fmt"
	"os"
	"text/template"
)

const htmlTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Build Dependency Graph</title>
    <script src="https://d3js.org/d3.v7.min.js"></script>
    <style>
        body {
            margin: 0;
            padding: 0;
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            background: #f5f5f5;
        }
        #container {
            width: 100vw;
            height: 100vh;
            overflow: hidden;
        }
        svg {
            width: 100%;
            height: 100%;
        }
        .node {
            cursor: pointer;
        }
        .node circle {
            stroke: #fff;
            stroke-width: 2px;
        }
        .node text {
            font-size: 11px;
            pointer-events: none;
        }
        .link {
            fill: none;
            stroke-opacity: 0.6;
        }
        .link.includes { stroke: #999; stroke-dasharray: 5,5; }
        .link.compiles_to { stroke: #2196F3; }
        .link.links_to { stroke: #4CAF50; stroke-width: 2px; }
        .link.depends_on { stroke: #FF9800; stroke-dasharray: 2,2; }

        .tooltip {
            position: absolute;
            background: rgba(0,0,0,0.8);
            color: white;
            padding: 8px 12px;
            border-radius: 4px;
            font-size: 12px;
            pointer-events: none;
            max-width: 300px;
            word-wrap: break-word;
        }

        .legend {
            position: absolute;
            top: 20px;
            right: 20px;
            background: white;
            padding: 15px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        .legend-item {
            display: flex;
            align-items: center;
            margin: 8px 0;
        }
        .legend-circle {
            width: 16px;
            height: 16px;
            border-radius: 50%;
            margin-right: 10px;
        }
        .legend-line {
            width: 30px;
            height: 2px;
            margin-right: 10px;
        }

        .controls {
            position: absolute;
            top: 20px;
            left: 20px;
            background: white;
            padding: 15px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        .controls button {
            display: block;
            width: 100%;
            padding: 8px 16px;
            margin: 5px 0;
            border: none;
            background: #2196F3;
            color: white;
            border-radius: 4px;
            cursor: pointer;
        }
        .controls button:hover {
            background: #1976D2;
        }

        .stats {
            position: absolute;
            bottom: 20px;
            left: 20px;
            background: white;
            padding: 15px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            font-size: 13px;
        }
    </style>
</head>
<body>
    <div id="container"></div>

    <div class="controls">
        <button onclick="resetZoom()">Reset View</button>
        <button onclick="toggleLabels()">Toggle Labels</button>
        <button onclick="centerGraph()">Center</button>
    </div>

    <div class="legend">
        <strong>Node Types</strong>
        <div class="legend-item">
            <div class="legend-circle" style="background: #a8d8ea;"></div>
            <span>Source</span>
        </div>
        <div class="legend-item">
            <div class="legend-circle" style="background: #d4f1f4;"></div>
            <span>Header</span>
        </div>
        <div class="legend-item">
            <div class="legend-circle" style="background: #ffcc80;"></div>
            <span>Object</span>
        </div>
        <div class="legend-item">
            <div class="legend-circle" style="background: #c5e1a5;"></div>
            <span>Library</span>
        </div>
        <div class="legend-item">
            <div class="legend-circle" style="background: #ef9a9a;"></div>
            <span>Executable</span>
        </div>
        <hr style="margin: 10px 0;">
        <strong>Edge Types</strong>
        <div class="legend-item">
            <div class="legend-line" style="background: #2196F3;"></div>
            <span>Compiles To</span>
        </div>
        <div class="legend-item">
            <div class="legend-line" style="background: #4CAF50; height: 3px;"></div>
            <span>Links To</span>
        </div>
        <div class="legend-item">
            <div class="legend-line" style="background: #999; border-top: 2px dashed #999; height: 0;"></div>
            <span>Includes</span>
        </div>
    </div>

    <div class="stats">
        <strong>Graph Statistics</strong><br>
        Nodes: <span id="node-count">0</span><br>
        Edges: <span id="edge-count">0</span>
    </div>

    <div class="tooltip" style="display: none;"></div>

    <script>
        const graphData = {{.GraphJSON}};

        // HTML escape function to prevent XSS
        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }

        // Convert graph data
        const nodes = Object.values(graphData.nodes).map(n => ({
            id: n.id,
            file: n.file,
            type: n.type,
            compiler: n.compiler || ''
        }));

        const links = graphData.edges.map(e => ({
            source: e.from,
            target: e.to,
            type: e.type
        }));

        // Update stats
        document.getElementById('node-count').textContent = nodes.length;
        document.getElementById('edge-count').textContent = links.length;

        // Color scale
        const colorScale = {
            'source': '#a8d8ea',
            'header': '#d4f1f4',
            'object': '#ffcc80',
            'library': '#c5e1a5',
            'executable': '#ef9a9a'
        };

        // Setup SVG
        const container = document.getElementById('container');
        const width = container.clientWidth;
        const height = container.clientHeight;

        const svg = d3.select('#container')
            .append('svg')
            .attr('viewBox', [0, 0, width, height]);

        const g = svg.append('g');

        // Zoom behavior
        const zoom = d3.zoom()
            .scaleExtent([0.1, 4])
            .on('zoom', (event) => g.attr('transform', event.transform));

        svg.call(zoom);

        // Arrow markers
        svg.append('defs').selectAll('marker')
            .data(['includes', 'compiles_to', 'links_to', 'depends_on'])
            .enter().append('marker')
            .attr('id', d => 'arrow-' + d)
            .attr('viewBox', '0 -5 10 10')
            .attr('refX', 20)
            .attr('refY', 0)
            .attr('markerWidth', 6)
            .attr('markerHeight', 6)
            .attr('orient', 'auto')
            .append('path')
            .attr('fill', d => {
                if (d === 'includes') return '#999';
                if (d === 'compiles_to') return '#2196F3';
                if (d === 'links_to') return '#4CAF50';
                return '#FF9800';
            })
            .attr('d', 'M0,-5L10,0L0,5');

        // Force simulation
        const simulation = d3.forceSimulation(nodes)
            .force('link', d3.forceLink(links).id(d => d.id).distance(100))
            .force('charge', d3.forceManyBody().strength(-300))
            .force('center', d3.forceCenter(width / 2, height / 2))
            .force('collision', d3.forceCollide().radius(30));

        // Links
        const link = g.append('g')
            .selectAll('line')
            .data(links)
            .enter().append('line')
            .attr('class', d => 'link ' + d.type)
            .attr('marker-end', d => 'url(#arrow-' + d.type + ')');

        // Node groups
        const node = g.append('g')
            .selectAll('.node')
            .data(nodes)
            .enter().append('g')
            .attr('class', 'node')
            .call(d3.drag()
                .on('start', dragstarted)
                .on('drag', dragged)
                .on('end', dragended));

        // Node circles
        node.append('circle')
            .attr('r', d => d.type === 'executable' ? 12 : d.type === 'library' ? 10 : 8)
            .attr('fill', d => colorScale[d.type] || '#ccc');

        // Node labels
        let showLabels = true;
        const labels = node.append('text')
            .attr('dx', 15)
            .attr('dy', 4)
            .text(d => {
                const name = d.file.split('/').pop();
                return name.length > 20 ? name.substring(0, 17) + '...' : name;
            });

        // Tooltip
        const tooltip = d3.select('.tooltip');

        node.on('mouseover', (event, d) => {
            tooltip.style('display', 'block')
                .html(` + "`" + `
                    <strong>${escapeHtml(d.file)}</strong><br>
                    Type: ${escapeHtml(d.type)}<br>
                    ${d.compiler ? 'Compiler: ' + escapeHtml(d.compiler) : ''}
                ` + "`" + `)
                .style('left', (event.pageX + 10) + 'px')
                .style('top', (event.pageY + 10) + 'px');
        })
        .on('mouseout', () => tooltip.style('display', 'none'));

        // Simulation tick
        simulation.on('tick', () => {
            link
                .attr('x1', d => d.source.x)
                .attr('y1', d => d.source.y)
                .attr('x2', d => d.target.x)
                .attr('y2', d => d.target.y);

            node.attr('transform', d => 'translate(' + d.x + ',' + d.y + ')');
        });

        // Drag functions
        function dragstarted(event, d) {
            if (!event.active) simulation.alphaTarget(0.3).restart();
            d.fx = d.x;
            d.fy = d.y;
        }

        function dragged(event, d) {
            d.fx = event.x;
            d.fy = event.y;
        }

        function dragended(event, d) {
            if (!event.active) simulation.alphaTarget(0);
            d.fx = null;
            d.fy = null;
        }

        // Control functions
        function resetZoom() {
            svg.transition().duration(750).call(
                zoom.transform,
                d3.zoomIdentity
            );
        }

        function toggleLabels() {
            showLabels = !showLabels;
            labels.style('display', showLabels ? 'block' : 'none');
        }

        function centerGraph() {
            const bounds = g.node().getBBox();
            const fullWidth = width;
            const fullHeight = height;
            const midX = bounds.x + bounds.width / 2;
            const midY = bounds.y + bounds.height / 2;
            const scale = 0.8 / Math.max(bounds.width / fullWidth, bounds.height / fullHeight);
            const translate = [fullWidth / 2 - scale * midX, fullHeight / 2 - scale * midY];

            svg.transition().duration(750).call(
                zoom.transform,
                d3.zoomIdentity.translate(translate[0], translate[1]).scale(scale)
            );
        }

        // Initial center
        setTimeout(centerGraph, 1000);
    </script>
</body>
</html>`

// RenderHTML renders the graph as an interactive HTML file with D3.js.
func RenderHTML(g *Graph, outputPath string) error {
	// Convert graph to JSON
	graphJSON, err := g.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to convert graph to JSON: %w", err)
	}

	// Parse template
	tmpl, err := template.New("graph").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Execute template
	data := struct {
		GraphJSON string
	}{
		GraphJSON: string(graphJSON),
	}

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// RenderDOT renders the graph as a DOT file for Graphviz.
func RenderDOT(g *Graph, outputPath string) error {
	dot := g.ToDOT()
	return os.WriteFile(outputPath, []byte(dot), 0644)
}

// RenderJSON renders the graph as a JSON file.
func RenderJSON(g *Graph, outputPath string) error {
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal graph: %w", err)
	}
	return os.WriteFile(outputPath, data, 0644)
}
