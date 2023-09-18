package cmd

import (
	"errors"

	"github.com/elliotchance/orderedmap/v2"
	"github.com/sirupsen/logrus"
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
			logrus.Debugf("Setting mapping coord for rootnode %s to %v", n.name, mappingCoord)
			g.nodes[n.index].mappingCoord = mappingCoord
			highestPositionPerLevel[0] = highestPositionPerLevel[0] + 1
		}
	}

	for _, n := range g.nodes {
		childLevel := n.mappingCoord.x + 1
		highestPosition := highestPositionPerLevel[childLevel]
		for _, child := range g.getChildren(n) {
			// Skip if the child already has a mapping coord
			if child.mappingCoord != nil {
				continue
			}

			// TODO: Change x/y depending on graph TD/LR. This is LR
			mappingCoord := g.reserveSpotInGrid(g.nodes[n.index], &coord{x: childLevel, y: highestPosition})
			logrus.Debugf("Setting mapping coord for child %s of parent %s to %v", child.name, n.name, mappingCoord)
			g.nodes[child.index].mappingCoord = mappingCoord
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
	for _, edge := range g.edges {
		g.drawEdge(edge)
	}
	return g.drawing
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
	parents := []*node{}
	for _, edge := range g.edges {
		if edge.to.name == n.name {
			parents = append(parents, edge.from)
		}
	}
	return parents
}
