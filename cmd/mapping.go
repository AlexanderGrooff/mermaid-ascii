package cmd

import (
	"github.com/elliotchance/orderedmap/v2"
	log "github.com/sirupsen/logrus"
)

type coord struct {
	x int
	y int
}

type node struct {
	name    string
	drawing drawing
	coord   coord
}

func (n *node) setCoord(c coord) {
	n.coord = c
}

type edge struct {
	from node
	to   node
	text string
}

type graph struct {
	nodes   []node
	edges   []edge
	drawing drawing
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

func (g graph) getChildren(n node) []node {
	edges := g.getEdgesFromNode(n)
	children := []node{}
	for _, edge := range edges {
		children = append(children, edge.to)
	}
	return children
}

func (g *graph) getOrCreateRootNode(name string, children []labeledChild) node {
	// Check if the node already exists.
	for _, existingRootNode := range g.nodes {
		if existingRootNode.name == name {
			log.Debug("Found existing root node ", existingRootNode.name, " at ", existingRootNode.coord)
			return existingRootNode
		}
	}
	parentCoord := g.positionNextRootNode(children)
	log.Debug("Creating new root node ", name, " at ", parentCoord)
	parentNode := node{name: name, drawing: drawBox(name), coord: parentCoord}
	g.drawNode(parentNode)
	g.appendNode(parentNode)
	return parentNode
}

func (g graph) positionNextRootNode(children []labeledChild) coord {
	if len(g.nodes) == 0 {
		return coord{x: 0, y: 0}
	}

	// Find existing child nodes to position the next root node next to.
	// This way we can take label width into account.
	var labelWidth int = 0
	for _, node := range g.nodes {
		for _, child := range children {
			if child.child == node.name {
				labelWidth = Max(len(child.label), labelWidth)
			}
		}
	}
	padding := Max(paddingBetweenX, labelWidth+4)

	w, _ := g.dimensions()
	return coord{x: w + padding, y: 0}
}

func (g *graph) getOrCreateChildNode(parent node, target labeledChild) node {
	// Check if the node already exists.
	for _, existingChildNode := range g.nodes {
		if existingChildNode.name == target.child {
			log.Debug("Found existing child node ", existingChildNode.name, " at ", existingChildNode.coord)
			return existingChildNode
		}
	}
	childNode := node{name: target.child, drawing: drawBox(target.child)}
	childCoord := g.findPositionChildNode(parent, childNode, target.label)
	childNode.setCoord(childCoord)
	g.drawNode(childNode)
	g.appendNode(childNode)
	log.Debug("Placed child node: ", childNode.coord)
	return childNode
}

func (g graph) findPositionChildNode(parent node, child node, arrowLabel string) coord {
	// Find a place to put the node, so it doesn't collide with any other nodes.
	// Place the node next to its parent node, if possible. Otherwise, place it
	// under the previous child node.
	parentWidth, _ := getDrawingSize(parent.drawing)

	// Add additional marging to the right of the parent node, so the arrow label
	// can be placed on the arrow to the child node.
	coordX := parent.coord.x + parentWidth + paddingBetweenX
	if arrowLabel != "" && paddingBetweenX < len(arrowLabel)+4 {
		coordX = parent.coord.x + parentWidth + len(arrowLabel) + 4
	}

	children := g.getChildren(parent)
	if len(children) == 0 {
		coordNextToParent := coord{coordX, parent.coord.y}
		log.Debug("Placing child node ", child.name, " next to parent node ", parent.name)
		return coordNextToParent
	}
	lastChildNode := children[len(children)-1]
	_, lastChildNodeHeight := getDrawingSize(lastChildNode.drawing)
	log.Debug("Placing child node ", child.name, " under last child node ", lastChildNode.name, " for parent node ", parent.name)
	return coord{x: lastChildNode.coord.x, y: lastChildNode.coord.y + lastChildNodeHeight + paddingBetweenY}
}

func (g graph) dimensions() (int, int) {
	return getDrawingSize(g.drawing)
}

func mkGraph(data *orderedmap.OrderedMap[string, []labeledChild]) graph {
	g := graph{drawing: mkDrawing(0, 0)}
	for el := data.Front(); el != nil; el = el.Next() {
		nodeName := el.Key
		children := el.Value
		parentNode := g.getOrCreateRootNode(nodeName, children)
		for _, target := range children {
			childNode := g.getOrCreateChildNode(parentNode, target)
			e := edge{from: parentNode, to: childNode, text: target.label}
			g.drawEdge(e)
			g.edges = append(g.edges, e)
		}
	}
	return g
}

func getArrowStartEndOffset(from node, to node) (coord, coord) {
	// Find which sides the arrow should start/end.
	// This is the middle of one of the sides, depending on the direction of the arrow.
	// Note that the coord returned is relative to the box.
	fromBoxWidth, fromBoxHeight := getDrawingSize(from.drawing)
	toBoxWidth, toBoxHeight := getDrawingSize(to.drawing)
	if from.coord.x == to.coord.x {
		// Vertical arrow
		if from.coord.y < to.coord.y {
			// Down
			return coord{fromBoxWidth / 2, fromBoxHeight}, coord{toBoxWidth / 2, 0}
		} else {
			// Up
			return coord{fromBoxWidth / 2, 0}, coord{toBoxWidth / 2, toBoxHeight}
		}
	} else if from.coord.y == to.coord.y {
		// Horizontal arrow
		if from.coord.x < to.coord.x {
			// Right
			return coord{fromBoxWidth, fromBoxHeight / 2}, coord{0, toBoxHeight / 2}
		} else {
			// Left
			return coord{0, fromBoxHeight / 2}, coord{toBoxWidth, toBoxHeight / 2}
		}
	} else {
		// Diagonal arrow
		if from.coord.x < to.coord.x {
			// Right
			if from.coord.y < to.coord.y {
				// Down
				return coord{fromBoxWidth / 2, fromBoxHeight}, coord{0, toBoxHeight / 2}
			} else {
				// Up
				return coord{fromBoxWidth / 2, 0}, coord{0, toBoxHeight / 2}
			}
		} else {
			// Left
			if from.coord.y < to.coord.y {
				// Down
				return coord{fromBoxWidth / 2, fromBoxHeight}, coord{toBoxWidth, toBoxHeight / 2}
			} else {
				// Up
				return coord{fromBoxWidth / 2, 0}, coord{toBoxWidth, toBoxHeight / 2}
			}
		}
	}
}
