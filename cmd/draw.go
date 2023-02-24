package cmd

import (
	"fmt"
	"strings"

	"github.com/elliotchance/orderedmap/v2"
	log "github.com/sirupsen/logrus"
)


type drawing [][]string

func (g *graph) drawNode(n node) {
	m := mergeDrawings(g.drawing, n.drawing, n.coord)
	g.drawing = m
}

func (g *graph) drawEdge(e edge) {
	arrowStart, arrowEnd := getArrowStartEndOffset(e.from, e.to)
	arrow := drawArrow(
		coord{x: e.from.coord.x + arrowStart.x, y: e.from.coord.y + arrowStart.y},
		coord{x: e.to.coord.x + arrowEnd.x, y: e.to.coord.y + arrowEnd.y},
	)
	m := mergeDrawings(g.drawing, arrow, coord{x: 0, y: 0})
	g.drawing = m
}

func drawMap(data *orderedmap.OrderedMap[string, []string]) {
	totalGraph := mkGraph(data)
	s := drawingToString(totalGraph.drawing)
	fmt.Println(s)
}

func drawBox(text string) drawing {
	from := coord{0, 0}
	// -1 because we start at 0
	to := coord{len(text) + boxBorderPadding*2 + boxBorderWidth*2 - 1, boxBorderWidth*2 + boxBorderPadding*2}
	boxDrawing := mkDrawing(Max(from.x, to.x), Max(from.y, to.y))
	log.Debug("Drawing box from ", from, " to ", to)
	// Draw top border
	for x := from.x; x < to.x; x++ {
		boxDrawing[x][from.y] = "-" // Horizontal line
	}
	// Draw bottom border
	for x := from.x; x < to.x; x++ {
		boxDrawing[x][to.y] = "-" // Horizontal line
	}
	// Draw left border
	for y := from.y; y < to.y; y++ {
		boxDrawing[from.x][y] = "|" // Vertical line
	}
	// Draw right border
	for y := from.y; y < to.y; y++ {
		boxDrawing[to.x][y] = "|" // Vertical line
	}
	// Draw text
	textY := from.y + boxBorderPadding + boxBorderWidth
	textXOffset := from.x + boxBorderPadding + boxBorderWidth
	for x := from.x + boxBorderPadding + boxBorderWidth; x < textXOffset+len(text); x++ {
		boxDrawing[x][textY] = string(text[x-textXOffset]) // Text
	}
	// Draw corners
	boxDrawing[from.x][from.y] = "+" // Top left corner
	boxDrawing[to.x][from.y] = "+"   // Top right corner
	boxDrawing[from.x][to.y] = "+"   // Bottom left corner
	boxDrawing[to.x][to.y] = "+"     // Bottom right corner

	return boxDrawing
}

func drawArrow(from coord, to coord) drawing {
	// Stop arrow one character before the end coord to stop just before the target
	arrowDrawing := mkDrawing(Max(from.x, to.x), Max(from.y, to.y))
	log.Debug("Drawing arrow from ", from, " to ", to)
	// Find the coord where the arrow should rotate
	rotateCoord := coord{from.x, to.y}
	// Draw from start to rotate
	if from.y <= rotateCoord.y {
		// Up
		for y := from.y + 1; y < rotateCoord.y; y++ {
			arrowDrawing[rotateCoord.x][y] = "|" // Vertical line
		}
	} else {
		// Down
		for y := rotateCoord.y; y < from.y; y++ {
			arrowDrawing[rotateCoord.x][y] = "|" // Vertical line
		}
	}
	// Draw from rotate to end
	if to.x >= rotateCoord.x {
		// Right
		for x := rotateCoord.x + 1; x < to.x; x++ {
			arrowDrawing[x][rotateCoord.y] = "-" // Horizontal line
		}
	} else {
		// Left
		for x := to.x + 1; x < rotateCoord.x; x++ {
			arrowDrawing[x][rotateCoord.y] = "-" // Horizontal line
		}
	}
	if from.x != to.x && from.y != to.y {
		arrowDrawing[rotateCoord.x][rotateCoord.y] = "+" // Corner
	}
	// Draw arrow head depending on direction
	if from.x == to.x {
		// Vertical arrow
		if from.y < to.y {
			// Down
			arrowDrawing[to.x][to.y-1] = "v"
		} else {
			// Up
			arrowDrawing[to.x][to.y] = "^"
		}
	} else if from.x < to.x {
		// Right
		arrowDrawing[to.x-1][to.y] = ">"
	} else {
		// Left
		arrowDrawing[to.x+1][to.y] = "<"
	}

	return arrowDrawing
}
func mergeDrawings(d1 drawing, d2 drawing, mergeCoord coord) drawing {
	maxX1, maxY1 := getDrawingSize(d1)
	maxX2, maxY2 := getDrawingSize(d2)
	maxX := Max(maxX1, maxX2+mergeCoord.x)
	maxY := Max(maxY1, maxY2+mergeCoord.y)
	mergedDrawing := mkDrawing(maxX, maxY)
	// Copy d1
	for x := 0; x <= maxX1; x++ {
		for y := 0; y <= maxY1; y++ {
			mergedDrawing[x][y] = d1[x][y]
		}
	}
	// Copy d2 with offset
	for x := 0; x <= maxX2; x++ {
		for y := 0; y <= maxY2; y++ {
			c := d2[x][y]
			if c != " " {
				mergedDrawing[x+mergeCoord.x][y+mergeCoord.y] = d2[x][y]
			}
		}
	}
	return mergedDrawing
}

func drawingToString(d drawing) string {
	maxX, maxY := getDrawingSize(d)
	dBuilder := strings.Builder{}
	for y := 0; y <= maxY; y++ {
		for x := 0; x <= maxX; x++ {
			dBuilder.WriteString(d[x][y])
		}
		if y != maxY {
			dBuilder.WriteString("\n")
		}
	}
	return dBuilder.String()
}

func mkDrawing(x int, y int) drawing {
	d := make(drawing, x+1)
	for i := 0; i <= x; i++ {
		d[i] = make([]string, y+1)
		for j := 0; j <= y; j++ {
			d[i][j] = " "
		}
	}
	return d
}

func getDrawingSize(d drawing) (int, int) {
	return len(d) - 1, len(d[0]) - 1
}

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
