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

func (g *graph) getOrCreateRootNode(name string) node {
	// Check if the node already exists.
	for _, existingRootNode := range g.nodes {
		if existingRootNode.name == name {
			log.Debug("Found existing root node ", existingRootNode.name, " at ", existingRootNode.coord)
			return existingRootNode
		}
	}
	parentCoord := g.positionNextRootNode()
	log.Debug("Creating new root node ", name, " at ", parentCoord)
	parentNode := node{name: name, drawing: drawBox(name), coord: parentCoord}
	g.drawNode(parentNode)
	g.appendNode(parentNode)
	return parentNode
}

func (g graph) positionNextRootNode() coord {
	if len(g.nodes) == 0 {
		return coord{x: 0, y: 0}
	}
	w, _ := g.dimensions()
	return coord{x: w + paddingBetweenX, y: 0}
}

func (g *graph) getOrCreateChildNode(parent node, name string) node {
	// Check if the node already exists.
	for _, existingChildNode := range g.nodes {
		if existingChildNode.name == name {
			log.Debug("Found existing child node ", existingChildNode.name, " at ", existingChildNode.coord)
			return existingChildNode
		}
	}
	childNode := node{name: name, drawing: drawBox(name)}
	childCoord := g.findPositionChildNode(parent, childNode)
	childNode.setCoord(childCoord)
	g.drawNode(childNode)
	g.appendNode(childNode)
	log.Debug("Placed child node: ", childNode.coord)
	return childNode
}

func (g graph) findPositionChildNode(parent node, child node) coord {
	// Find a place to put the node, so it doesn't collide with any other nodes.
	// Place the node next to its parent node, if possible. Otherwise, place it
	// under the previous child node.
	parentWidth, _ := getDrawingSize(parent.drawing)

	// Check if the child node can be placed next to the parent node.
	coordNextToParent := coord{parent.coord.x + parentWidth + paddingBetweenX, parent.coord.y}
	if !doDrawingsCollide(g.drawing, child.drawing, coordNextToParent) {
		log.Debug("Placing child node ", child.name, " next to parent node ", parent.name)
		return coordNextToParent
	} else {
		// The child node can't be placed next to the parent node.
		// Find the last child node, and place the node under that one.
		// If there are no child nodes, place it under the parent node.
		children := g.getChildren(parent)
		if len(children) == 0 {
			log.Debug("Couldn't find position for child node ", child.name, " for parent node ", parent.name)
			return coord{x: 15, y: 15}
		}
		lastChildNode := children[len(children)-1]
		_, lastChildNodeHeight := getDrawingSize(lastChildNode.drawing)
		log.Debug("Placing child node ", child.name, " under last child node ", lastChildNode.name, " for parent node ", parent.name)
		return coord{x: lastChildNode.coord.x, y: lastChildNode.coord.y + lastChildNodeHeight + paddingBetweenY}
	}
}

func (g graph) dimensions() (int, int) {
	return getDrawingSize(g.drawing)
}

func mkGraph(data *orderedmap.OrderedMap[string, []labeledChild]) graph {
	g := graph{drawing: mkDrawing(0, 0)}
	for el := data.Front(); el != nil; el = el.Next() {
		nodeName := el.Key
		children := el.Value
		parentNode := g.getOrCreateRootNode(nodeName)
		for _, child := range children {
			childNode := g.getOrCreateChildNode(parentNode, child.child)
			e := edge{from: parentNode, to: childNode, text: child.label}
			g.drawEdge(e)
			g.edges = append(g.edges, e)
		}
	}
	return g
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
