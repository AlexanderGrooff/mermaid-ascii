package cmd

import (
	"errors"

	"github.com/elliotchance/orderedmap/v2"
)

type coord struct {
	x int
	y int
}

type graph struct {
	nodes       []*node
	edges       []*edge
	drawing     *drawing
	grid        map[coord]*node
	columnWidth map[int]int
}

func (g graph) dimensions() (int, int) {
	return getDrawingSize(g.drawing)
}

func mkGraph(data *orderedmap.OrderedMap[string, []string]) graph {
	g := graph{drawing: mkDrawing(0, 0)}
	g.grid = make(map[coord]*node)
	g.columnWidth = make(map[int]int)
	index := 0
	for el := data.Front(); el != nil; el = el.Next() {
		nodeName := el.Key
		children := el.Value
		// Get or create parent node
		parentNode, err := g.getNode(nodeName)
		if err != nil {
			parentNode = &node{name: nodeName, index: index}
			g.appendNode(parentNode)
			index += 1
		}
		for _, childNodeName := range children {
			childNode, err := g.getNode(childNodeName)
			if err != nil {
				childNode = &node{name: childNodeName, index: index}
				g.appendNode(childNode)
				index += 1
			}
			e := edge{from: parentNode, to: childNode, text: ""}
			g.edges = append(g.edges, &e)
		}
	}

	g.createMapping()
	return g
}

func (g *graph) createMapping() {
	// Set mapping coord for every node in the graph
	highestPositionPerLevel := []int{}
	// Init array with 0 values
	// TODO: I'm sure there's a better way of doing this
	for i := 0; i < 100; i++ {
		highestPositionPerLevel = append(highestPositionPerLevel, 0)
	}

	// Set root nodes to level 0
	for _, n := range g.nodes {
		if len(g.getParents(n)) == 0 {
			// TODO: Change x/y depending on graph TD/LR. This is LR
			mappingCoord := g.reserveSpotInGrid(g.nodes[n.index], &coord{x: 0, y: highestPositionPerLevel[0]})
			g.nodes[n.index].mappingCoord = mappingCoord
			highestPositionPerLevel[0] = highestPositionPerLevel[0] + 1
		}
	}

	for _, n := range g.nodes {
		for _, child := range g.getChildren(n) {
			// TODO: Change x/y depending on graph TD/LR. This is LR
			childLevel := n.mappingCoord.x + 1
			highestPosition := highestPositionPerLevel[childLevel]
			g.nodes[child.index].mappingCoord = &coord{x: childLevel, y: highestPosition}
			highestPositionPerLevel[childLevel] = highestPosition + 1
		}
	}

	// After mapping coords are set, set drawing coords
	for _, n := range g.nodes {
		g.nodes[n.index].setCoord(g.mappingToDrawingCoord(n))
		g.nodes[n.index].setDrawing()
	}
}

func (g *graph) draw() *drawing {
	// Draw all nodes.
	for _, node := range g.nodes {
		if !node.drawn {
			g.drawNode(node)
		}
	}
	return g.drawing
}

func doDrawingsCollide(drawing1 *drawing, drawing2 *drawing, offset coord) bool {
	// Check if any of the drawing2 characters overlap with drawing1 characters.
	// The offset is the coord of drawing2 relative to drawing1.
	drawing1Width, drawing1Height := getDrawingSize(drawing1)
	drawing2Width, drawing2Height := getDrawingSize(drawing2)
	for x := 0; x < drawing2Width; x++ {
		for y := 0; y < drawing2Height; y++ {
			// Check if drawing2[x][y] overlaps with drawing1[x+offset.x][y+offset.y]
			if x+offset.x >= 0 && x+offset.x < drawing1Width &&
				y+offset.y >= 0 && y+offset.y < drawing1Height &&
				(*drawing2)[x][y] != " " &&
				(*drawing1)[x+offset.x][y+offset.y] != " " {
				return true
			}
		}
	}
	return false
}

func (g *graph) getNode(nodeName string) (*node, error) {
	for _, n := range g.nodes {
		if n.name == nodeName {
			return n, nil
		}
	}
	return &node{}, errors.New("node " + nodeName + " not found")
}

func (g *graph) appendNode(n *node) {
	g.nodes = append(g.nodes, n)
}

func (g graph) getEdgesFromNode(n *node) []edge {
	edges := []edge{}
	for _, edge := range g.edges {
		if (edge.from.name) == (n.name) {
			edges = append(edges, *edge)
		}
	}
	return edges
}

func (g *graph) getChildren(n *node) []*node {
	edges := g.getEdgesFromNode(n)
	children := []*node{}
	for _, edge := range edges {
		if edge.from.name == n.name {
			children = append(children, edge.to)
		}
	}
	return children
}

func (g *graph) getParents(n *node) []*node {
	edges := g.getEdgesFromNode(n)
	parents := []*node{}
	for _, edge := range edges {
		if edge.to.name == n.name {
			parents = append(parents, edge.from)
		}
	}
	return parents
}
