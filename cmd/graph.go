package cmd

import (
	"github.com/elliotchance/orderedmap/v2"
)

type coord struct {
	x int
	y int
}

type graph struct {
	nodes   []node
	edges   []edge
	drawing drawing
}

func (g graph) dimensions() (int, int) {
	return getDrawingSize(g.drawing)
}

func mkGraph(data *orderedmap.OrderedMap[string, []string]) graph {
	g := graph{drawing: mkDrawing(0, 0)}
	index := 0
	for el := data.Front(); el != nil; el = el.Next() {
		nodeName := el.Key
		children := el.Value
		parentNode := &node{name: nodeName, index: index}
		if !g.isNodeInGraph(*parentNode) {
			g.appendNode(*parentNode)
			index += 1
		}
		for _, childNodeName := range children {
			childNode := &node{name: childNodeName, index: index}
			if !g.isNodeInGraph(*childNode) {
				g.appendNode(*childNode)
				index += 1
			}
			e := edge{from: *parentNode, to: *childNode, text: ""}
			g.edges = append(g.edges, e)
		}
	}

	// Set root node
	rootNode := &g.nodes[0]
	rootNode.setCoord(&coord{x: 0, y: 0})
	rootNode.level = 1

	return g
}

func (g *graph) createMapping() {
	highestPositionPerLevel := []int{}
	// Init array with 0 values
	// TODO: I'm sure there's a better way of doing this
	for i := 0; i < 100; i++ {
		highestPositionPerLevel = append(highestPositionPerLevel, 0)
	}
	for _, n := range g.nodes {
		for _, child := range g.getChildren(n) {
			g.nodes[child.index].level = n.level + 1
			// TODO: Change x/y depending on graph TD/LR. This is LR
			g.nodes[child.index].setCoord(&coord{x: n.level + 1, y: highestPositionPerLevel[n.level+1]})
			highestPositionPerLevel[n.level+1] = highestPositionPerLevel[n.level+1] + 1
		}
	}
}

func (g *graph) draw() drawing {
	// Ensure all nodes are mapped within the graph
	g.createMapping()

	// Draw all nodes.
	for _, node := range g.nodes {
		if !node.drawn {
			g.drawNode(node)
		}
	}
	return g.drawing
}

func doDrawingsCollide(drawing1 drawing, drawing2 drawing, offset coord) bool {
	// Check if any of the drawing2 characters overlap with drawing1 characters.
	// The offset is the coord of drawing2 relative to drawing1.
	drawing1Width, drawing1Height := getDrawingSize(drawing1)
	drawing2Width, drawing2Height := getDrawingSize(drawing2)
	for x := 0; x < drawing2Width; x++ {
		for y := 0; y < drawing2Height; y++ {
			// Check if drawing2[x][y] overlaps with drawing1[x+offset.x][y+offset.y]
			if x+offset.x >= 0 && x+offset.x < drawing1Width &&
				y+offset.y >= 0 && y+offset.y < drawing1Height &&
				drawing2[x][y] != " " &&
				drawing1[x+offset.x][y+offset.y] != " " {
				return true
			}
		}
	}
	return false
}

func (g *graph) isNodeInGraph(node node) bool {
	for _, n := range g.nodes {
		if n.name == node.name {
			return true
		}
	}
	return false
}

func (g *graph) appendNode(n node) {
	g.nodes = append(g.nodes, n)
}

func (g graph) getEdgesFromNode(node node) []edge {
	edges := []edge{}
	for _, edge := range g.edges {
		if (edge.from.name) == (node.name) {
			edges = append(edges, edge)
		}
	}
	return edges
}

func (g *graph) getChildren(n node) []node {
	edges := g.getEdgesFromNode(n)
	children := []node{}
	for _, edge := range edges {
		if edge.from.name == n.name {
			children = append(children, edge.to)
		}
	}
	return children
}
