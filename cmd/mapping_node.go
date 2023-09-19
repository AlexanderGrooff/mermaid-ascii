package cmd

import "github.com/sirupsen/logrus"

type node struct {
	name         string
	drawing      *drawing
	drawingCoord *coord
	mappingCoord *coord
	drawn        bool
	index        int // Index of the node in the graph.nodes slice
}

func (n *node) setCoord(c *coord) {
	n.drawingCoord = c
}

func (n *node) setDrawing() *drawing {
	d := drawBox(n.name)
	n.drawing = d
	return d
}

func (g *graph) mappingToDrawingCoord(n *node) *coord {
	// For every node there is:
	// - 2 lines of border
	// - 1 line of text
	// - 2x padding
	// - 2x margin
	w := 2*boxBorderWidth + 2*boxBorderPadding + len(n.name)

	labelLength := 0
	for _, e := range g.getEdgesFromNode(n) {
		// Does the edge go to the next column?
		if e.to.mappingCoord.x == n.mappingCoord.x+1 {
			labelLength = Max(labelLength, len(e.text))
		}
	}

	// Next to that you have previous columns, which have a max width based on the longest name
	prevX := 0
	for i := 0; i < n.mappingCoord.x; i++ {
		prevX += g.columnWidth[i] + 2*paddingBetweenX
	}
	g.columnWidth[n.mappingCoord.x] = Max(g.columnWidth[n.mappingCoord.x], w+labelLength)

	x := prevX
	y := n.mappingCoord.y * 2 * paddingBetweenY
	drawingCoord := coord{x: x, y: y}
	logrus.Debugf("Mapping coord for %s from %v to %v", n.name, *n.mappingCoord, drawingCoord)
	return &drawingCoord
}

func (g *graph) reserveSpotInGrid(n *node, requestedCoord *coord) *coord {
	if g.grid[*requestedCoord] != nil {
		logrus.Debugf("Coord %d,%d is already taken", requestedCoord.x, requestedCoord.y)
		// TODO: Change this based on TD/LR
		return g.reserveSpotInGrid(n, &coord{x: requestedCoord.x, y: requestedCoord.y + 1})
	}
	g.grid[*requestedCoord] = n
	return requestedCoord
}
