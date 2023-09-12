package cmd

import (
	log "github.com/sirupsen/logrus"
)

type node struct {
	name    string
	drawing drawing
	coord   *coord // Coord of the node within the grid
	drawn   bool
	level   int // How many layers down is this node from the root node? Root node level is 1. (0 is default value for unset integers)
	index   int // Index of the node in the graph.nodes slice
}

func (n *node) setCoord(c *coord) {
	n.coord = c
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
	parentNode := node{name: name, drawing: drawBox(name), coord: &parentCoord}
	// g.drawNode(parentNode)
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

func (g graph) findPositionChildNode(parent node, child node) *coord {
	// Find a place to put the node, so it doesn't collide with any other nodes.
	// Place the node next to its parent node, if possible. Otherwise, place it
	// under the previous child node.
	parentWidth, _ := getDrawingSize(parent.drawing)

	// Check if the child node can be placed next to the parent node.
	coordNextToParent := coord{parent.coord.x + parentWidth + paddingBetweenX, parent.coord.y}
	if !doDrawingsCollide(g.drawing, child.drawing, coordNextToParent) {
		log.Debug("Placing child node ", child.name, " next to parent node ", parent.name)
		return &coordNextToParent
	} else {
		// The child node can't be placed next to the parent node.
		// Find the last child node, and place the node under that one.
		// If there are no child nodes, place it under the parent node.
		children := g.getChildren(parent)
		if len(children) == 0 {
			log.Debug("Couldn't find position for child node ", child.name, " for parent node ", parent.name)
			return &coord{x: 15, y: 15}
		}
		lastChildNode := children[len(children)-1]
		_, lastChildNodeHeight := getDrawingSize(lastChildNode.drawing)
		log.Debug("Placing child node ", child.name, " under last child node ", lastChildNode.name, " for parent node ", parent.name)
		return &coord{x: lastChildNode.coord.x, y: lastChildNode.coord.y + lastChildNodeHeight + paddingBetweenY}
	}
}

func (n node) draw() drawing {
	// TODO: convert coords to drawing coords and draw the thing
	return n.drawing
}
