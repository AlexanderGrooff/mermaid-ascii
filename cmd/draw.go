package cmd

import (
	"fmt"
	"strings"

	"github.com/elliotchance/orderedmap/v2"
	log "github.com/sirupsen/logrus"
)

type drawing [][]string

func (g *graph) drawNode(n *node) {
	m := mergeDrawings(g.drawing, n.drawing, *n.drawingCoord)
	g.drawing = m
}

func (g *graph) drawEdge(e *edge) {
	arrowStart, arrowEnd := getArrowStartEndOffset(e.from, e.to)
	arrowFrom := coord{x: e.from.drawingCoord.x + arrowStart.x, y: e.from.drawingCoord.y + arrowStart.y}
	arrowTo := coord{x: e.to.drawingCoord.x + arrowEnd.x, y: e.to.drawingCoord.y + arrowEnd.y}
	g.drawing.drawArrow(
		arrowFrom,
		arrowTo,
		e.text,
	)
}

func (d *drawing) drawText(start coord, text string) {
	// Increase dimensions if necessary.
	d.increaseSize(start.x+len(text), start.y)
	log.Debug("Drawing '", text, "' from ", start, " to ", coord{x: start.x + len(text), y: start.y})
	for x := 0; x < len(text); x++ {
		(*d)[x+start.x][start.y] = string(text[x])
	}
}

func (d *drawing) drawLine(from coord, to coord, offsetFrom int, offsetTo int) {
	// Offset determines how far from the actual coord the line should start/stop.
	direction := determineDirection(from, to)
	log.Debug("Drawing line from ", from, " to ", to, " direction: ", direction, " offsetFrom: ", offsetFrom, " offsetTo: ", offsetTo)
	switch direction {
	case Up:
		for y := from.y - offsetFrom; y >= to.y-offsetTo; y-- {
			(*d)[from.x][y] = "|"
		}
	case Down:
		for y := from.y + offsetFrom; y <= to.y+offsetTo; y++ {
			(*d)[from.x][y] = "|"
		}
	case Left:
		for x := from.x - offsetFrom; x >= to.x-offsetTo; x-- {
			(*d)[x][from.y] = "-"
		}
	case Right:
		for x := from.x + offsetFrom; x <= to.x+offsetTo; x++ {
			(*d)[x][from.y] = "-"
		}
	case UpperLeft:
		for x, y := from.x, from.y-offsetFrom; x >= to.x-offsetTo && y >= to.y-offsetTo; x, y = x-1, y-1 {
			(*d)[x][y] = "\\"
		}
	case UpperRight:
		for x, y := from.x, from.y-offsetFrom; x <= to.x+offsetTo && y >= to.y-offsetTo; x, y = x+1, y-1 {
			(*d)[x][y] = "/"
		}
	case LowerLeft:
		for x, y := from.x, from.y+offsetFrom; x >= to.x-offsetTo && y <= to.y+offsetTo; x, y = x-1, y+1 {
			(*d)[x][y] = "/"
		}
	case LowerRight:
		for x, y := from.x, from.y+offsetFrom; x <= to.x+offsetTo && y <= to.y+offsetTo; x, y = x+1, y+1 {
			(*d)[x][y] = "\\"
		}
	}
}

func drawMap(data *orderedmap.OrderedMap[string, []labeledChild]) {
	g := mkGraph(data)
	s := drawingToString(g.draw())
	fmt.Println(s)
}

func drawBox(text string) *drawing {
	from := coord{0, 0}
	// -1 because we start at 0
	to := coord{len(text) + boxBorderPadding*2 + boxBorderWidth*2 - 1, boxBorderWidth*2 + boxBorderPadding*2}
	boxDrawing := *(mkDrawing(Max(from.x, to.x), Max(from.y, to.y)))
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

	return &boxDrawing
}

func (d *drawing) increaseSize(x int, y int) {
	currSizeX, currSizeY := getDrawingSize(d)
	drawingWithNewSize := mkDrawing(Max(x, currSizeX), Max(y, currSizeY))
	*d = *mergeDrawings(drawingWithNewSize, d, coord{0, 0})
}

func mergeDrawings(d1 *drawing, d2 *drawing, mergeCoord coord) *drawing {
	maxX1, maxY1 := getDrawingSize(d1)
	maxX2, maxY2 := getDrawingSize(d2)
	maxX := Max(maxX1, maxX2+mergeCoord.x)
	maxY := Max(maxY1, maxY2+mergeCoord.y)
	mergedDrawing := mkDrawing(maxX, maxY)
	// Copy d1
	for x := 0; x <= maxX1; x++ {
		for y := 0; y <= maxY1; y++ {
			(*mergedDrawing)[x][y] = (*d1)[x][y]
		}
	}
	// Copy d2 with offset
	for x := 0; x <= maxX2; x++ {
		for y := 0; y <= maxY2; y++ {
			c := (*d2)[x][y]
			if c != " " {
				(*mergedDrawing)[x+mergeCoord.x][y+mergeCoord.y] = (*d2)[x][y]
			}
		}
	}
	return mergedDrawing
}

func drawingToString(d *drawing) string {
	maxX, maxY := getDrawingSize(d)
	dBuilder := strings.Builder{}
	for y := 0; y <= maxY; y++ {
		for x := 0; x <= maxX; x++ {
			dBuilder.WriteString((*d)[x][y])
		}
		if y != maxY {
			dBuilder.WriteString("\n")
		}
	}
	return dBuilder.String()
}

func mkDrawing(x int, y int) *drawing {
	d := make(drawing, x+1)
	for i := 0; i <= x; i++ {
		d[i] = make([]string, y+1)
		for j := 0; j <= y; j++ {
			d[i][j] = " "
		}
	}
	return &d
}

func getDrawingSize(d *drawing) (int, int) {
	return len(*d) - 1, len((*d)[0]) - 1
}

func determineDirection(from coord, to coord) direction {
	if from.x == to.x {
		if from.y < to.y {
			return Down
		} else {
			return Up
		}
	} else if from.y == to.y {
		if from.x < to.x {
			return Right
		} else {
			return Left
		}
	} else if from.x < to.x {
		if from.y < to.y {
			return LowerRight
		} else {
			return UpperRight
		}
	} else {
		if from.y < to.y {
			return LowerLeft
		} else {
			return UpperLeft
		}
	}
}
