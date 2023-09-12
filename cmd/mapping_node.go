package cmd

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

func (n node) draw() drawing {
	// TODO: convert coords to drawing coords and draw the thing
	return drawBox(n.name)
}
