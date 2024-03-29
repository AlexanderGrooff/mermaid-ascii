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
	nodes        []*node
	edges        []*edge
	drawing      *drawing
	grid         map[coord]*node
	columnWidth  map[int]int
	styleClasses map[string]styleClass
}

func mkGraph(data *orderedmap.OrderedMap[string, []textEdge]) graph {
	g := graph{drawing: mkDrawing(0, 0)}
	g.grid = make(map[coord]*node)
	g.columnWidth = make(map[int]int)
	g.styleClasses = make(map[string]styleClass)

	index := 0
	for el := data.Front(); el != nil; el = el.Next() {
		nodeName := el.Key
		children := el.Value
		// Get or create parent node
		parentNode, err := g.getNode(nodeName)
		if err != nil {
			parentNode = &node{name: nodeName, index: index, styleClassName: ""}
			g.appendNode(parentNode)
			index += 1
		}
		for _, textEdge := range children {
			childNode, err := g.getNode(textEdge.child.name)
			if err != nil {
				childNode = &node{name: textEdge.child.name, index: index, styleClassName: textEdge.child.styleClass}
				parentNode.styleClassName = textEdge.parent.styleClass
				g.appendNode(childNode)
				index += 1
			}
			e := edge{from: parentNode, to: childNode, text: textEdge.label}
			g.edges = append(g.edges, &e)
		}
	}
	return g
}

func (g *graph) setStyleClasses(styleClasses map[string]styleClass) {
	logrus.Debugf("Setting style classes to %v", styleClasses)
	g.styleClasses = styleClasses
	for _, n := range g.nodes {
		if n.styleClassName != "" {
			logrus.Debugf("Setting style class for node %s to %s", n.name, n.styleClassName)
			(*n).styleClass = g.styleClasses[n.styleClassName]
		}
	}
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
			var mappingCoord *coord
			if graphDirection == "LR" {
				mappingCoord = g.reserveSpotInGrid(g.nodes[n.index], &coord{x: 0, y: highestPositionPerLevel[0]})
			} else {
				mappingCoord = g.reserveSpotInGrid(g.nodes[n.index], &coord{x: highestPositionPerLevel[0], y: 0})
			}
			logrus.Debugf("Setting mapping coord for rootnode %s to %v", n.name, mappingCoord)
			g.nodes[n.index].mappingCoord = mappingCoord
			highestPositionPerLevel[0] = highestPositionPerLevel[0] + 1
		}
	}

	for _, n := range g.nodes {
		var childLevel int
		if graphDirection == "LR" {
			childLevel = n.mappingCoord.x + 1
		} else {
			childLevel = n.mappingCoord.y + 1
		}
		highestPosition := highestPositionPerLevel[childLevel]
		for _, child := range g.getChildren(n) {
			// Skip if the child already has a mapping coord
			if child.mappingCoord != nil {
				continue
			}

			var mappingCoord *coord
			if graphDirection == "LR" {
				mappingCoord = g.reserveSpotInGrid(g.nodes[n.index], &coord{x: childLevel, y: highestPosition})
			} else {
				mappingCoord = g.reserveSpotInGrid(g.nodes[n.index], &coord{x: highestPosition, y: childLevel})
			}
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
