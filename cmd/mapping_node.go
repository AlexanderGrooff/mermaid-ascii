package cmd

import "github.com/sirupsen/logrus"

type node struct {
	name           string
	drawing        *drawing
	drawingCoord   *drawingCoord
	gridCoord      *gridCoord
	drawn          bool
	index          int // Index of the node in the graph.nodes slice
	styleClassName string
	styleClass     styleClass
}

func (n *node) setCoord(c *drawingCoord) {
	n.drawingCoord = c
}

func (n *node) setDrawing() *drawing {
	d := drawBox(n)
	n.drawing = d
	return d
}

func (g *graph) mappingToDrawingCoord(n *node) *drawingCoord {
	// For every node there is:
	// - 2 lines of border
	// - 1 line of text
	// - 2x padding
	// - 2x margin
	w := 2*boxBorderWidth + 2*boxBorderPadding + len(n.name)

	labelLength := 0
	for _, e := range g.getEdgesFromNode(n) {
		// Does the edge go to the next column?
		if e.to.gridCoord.x == n.gridCoord.x+1 {
			labelLength = Max(labelLength, len(e.text))
		}
	}

	// Next to that you have previous columns, which have a max width based on the longest name
	prevX := 0
	for i := 0; i < n.gridCoord.x; i++ {
		prevX += g.columnWidth[i] + 2*paddingBetweenX
	}
	g.columnWidth[n.gridCoord.x] = Max(g.columnWidth[n.gridCoord.x], w+labelLength)

	x := prevX
	y := n.gridCoord.y * 2 * paddingBetweenY
	drawingCoord := drawingCoord{x: x, y: y}
	logrus.Debugf("Mapping coord for %s from %v to %v", n.name, *n.gridCoord, drawingCoord)
	return &drawingCoord
}

func (g *graph) reserveSpotInGrid(n *node, requestedCoord *gridCoord) *gridCoord {
	if g.grid[*requestedCoord] != nil {
		logrus.Debugf("Coord %d,%d is already taken", requestedCoord.x, requestedCoord.y)
		if graphDirection == "LR" {
			return g.reserveSpotInGrid(n, &gridCoord{x: requestedCoord.x, y: requestedCoord.y + 1})
		} else {
			return g.reserveSpotInGrid(n, &gridCoord{x: requestedCoord.x + 1, y: requestedCoord.y})
		}
	}
	g.grid[*requestedCoord] = n
	return requestedCoord
}
