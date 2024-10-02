package cmd

import (
	"fmt"
	"strings"

	orderedmap "github.com/elliotchance/orderedmap/v2"
	"github.com/gookit/color"
	log "github.com/sirupsen/logrus"
)

type drawing [][]string

type styleClass struct {
	name   string
	styles map[string]string
}

func (g *graph) drawNode(n *node) {
	log.Debug("Drawing node ", n.name, " at ", *n.drawingCoord)
	m := mergeDrawings(g.drawing, n.drawing, *n.drawingCoord)
	g.drawing = m
}

func (g *graph) drawEdge(e *edge) {
	dir, oppositeDir := determineStartAndEndDir(e)
	from := e.from.gridCoord.Direction(dir)
	to := e.to.gridCoord.Direction(oppositeDir)
	log.Debugf("Drawing edge between %v (direction %v) and %v (direction %v)", *e.from, dir, *e.to, oppositeDir)
	g.drawArrow(from, to, e.text)
}

func (d *drawing) drawText(start drawingCoord, text string) {
	// Increase dimensions if necessary.
	d.increaseSize(start.x+len(text), start.y)
	log.Debug("Drawing '", text, "' from ", start, " to ", drawingCoord{x: start.x + len(text), y: start.y})
	for x := 0; x < len(text); x++ {
		(*d)[x+start.x][start.y] = string(text[x])
	}
}

func (d *drawing) drawLine(from drawingCoord, to drawingCoord, offsetFrom int, offsetTo int) []drawingCoord {
	// Offset determines how far from the actual coord the line should start/stop.
	direction := determineDirection(genericCoord(from), genericCoord(to))
	drawnCoords := make([]drawingCoord, 0)
	log.Debug("Drawing line from ", from, " to ", to, " direction: ", direction, " offsetFrom: ", offsetFrom, " offsetTo: ", offsetTo)
	switch direction {
	case Up:
		for y := from.y - offsetFrom; y >= to.y-offsetTo; y-- {
			drawnCoords = append(drawnCoords, drawingCoord{from.x, y})
			(*d)[from.x][y] = "|"
		}
	case Down:
		for y := from.y + offsetFrom; y <= to.y+offsetTo; y++ {
			drawnCoords = append(drawnCoords, drawingCoord{from.x, y})
			(*d)[from.x][y] = "|"
		}
	case Left:
		for x := from.x - offsetFrom; x >= to.x-offsetTo; x-- {
			drawnCoords = append(drawnCoords, drawingCoord{x, from.y})
			(*d)[x][from.y] = "-"
		}
	case Right:
		for x := from.x + offsetFrom; x <= to.x+offsetTo; x++ {
			drawnCoords = append(drawnCoords, drawingCoord{x, from.y})
			(*d)[x][from.y] = "-"
		}
	case UpperLeft:
		for x, y := from.x, from.y-offsetFrom; x >= to.x-offsetTo && y >= to.y-offsetTo; x, y = x-1, y-1 {
			drawnCoords = append(drawnCoords, drawingCoord{x, y})
			(*d)[x][y] = "\\"
		}
	case UpperRight:
		for x, y := from.x, from.y-offsetFrom; x <= to.x+offsetTo && y >= to.y-offsetTo; x, y = x+1, y-1 {
			drawnCoords = append(drawnCoords, drawingCoord{x, y})
			(*d)[x][y] = "/"
		}
	case LowerLeft:
		for x, y := from.x, from.y+offsetFrom; x >= to.x-offsetTo && y <= to.y+offsetTo; x, y = x-1, y+1 {
			drawnCoords = append(drawnCoords, drawingCoord{x, y})
			(*d)[x][y] = "/"
		}
	case LowerRight:
		for x, y := from.x, from.y+offsetFrom; x <= to.x+offsetTo && y <= to.y+offsetTo; x, y = x+1, y+1 {
			drawnCoords = append(drawnCoords, drawingCoord{x, y})
			(*d)[x][y] = "\\"
		}
	}
	return drawnCoords
}

func drawMap(data *orderedmap.OrderedMap[string, []textEdge], styleClasses map[string]styleClass) string {
	g := mkGraph(data)
	g.setStyleClasses(styleClasses)
	g.createMapping()
	d := g.draw()
	if Verbose {
		d = d.debugDrawingWrapper()
		d = d.debugCoordWrapper(g)
	}
	s := drawingToString(d)
	fmt.Println(s)
	return s
}

func drawBox(n *node) *drawing {
	from := drawingCoord{0, 0}
	// -1 because we start at 0
	to := drawingCoord{len(n.name) + boxBorderPadding*2 + boxBorderWidth*2 - 1, boxBorderWidth*2 + boxBorderPadding*2}
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
	// Set up text color
	var c color.RGBColor
	log.Debugf("Color for node %s is %s", n.name, n.styleClass)
	if n.styleClass.styles["color"] != "" {
		c = color.HEX(n.styleClass.styles["color"])
	}
	// Draw text
	textY := from.y + boxBorderPadding + boxBorderWidth
	textXOffset := from.x + boxBorderPadding + boxBorderWidth
	for x := from.x + boxBorderPadding + boxBorderWidth; x < textXOffset+len(n.name); x++ {
		if n.styleClass.styles["color"] != "" {
			log.Debugf("Setting color for node %s to %s", n.name, n.styleClass.styles["color"])
			boxDrawing[x][textY] = c.Sprint(string(n.name[x-textXOffset])) // Text
		} else {
			boxDrawing[x][textY] = string(n.name[x-textXOffset]) // Text
		}
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
	*d = *mergeDrawings(drawingWithNewSize, d, drawingCoord{0, 0})
}

func mergeDrawings(d1 *drawing, d2 *drawing, mergeCoord drawingCoord) *drawing {
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

func copyCanvas(toBeCopied *drawing) *drawing {
	x, y := getDrawingSize(toBeCopied)
	return mkDrawing(x, y)
}

func getDrawingSize(d *drawing) (int, int) {
	return len(*d) - 1, len((*d)[0]) - 1
}

func determineDirection(from genericCoord, to genericCoord) direction {
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

func (d drawing) debugDrawingWrapper() *drawing {
	maxX, maxY := getDrawingSize(&d)
	debugDrawing := *mkDrawing(maxX+2, maxY+1)
	for x := 0; x <= maxX; x++ {
		debugDrawing[x+2][0] = fmt.Sprintf("%d", x%10)
	}
	for y := 0; y <= maxY; y++ {
		debugDrawing[0][y+1] = fmt.Sprintf("%2d", y)
	}

	return mergeDrawings(&debugDrawing, &d, drawingCoord{1, 1})
}

func (d drawing) debugCoordWrapper(g graph) *drawing {
	maxX, maxY := getDrawingSize(&d)
	debugDrawing := *mkDrawing(maxX+2, maxY+1)
	currX := 3
	currY := 2
	for x := 0; currX <= maxX+g.columnWidth[x]; x++ {
		w := g.columnWidth[x]
		debugPos := currX + w/2
		// log.Debugf("Grid coord %d has width %d: %d", x, w, currX)
		debugDrawing[debugPos][0] = fmt.Sprintf("%d", x%10)
		currX += w
	}
	for y := 0; currY <= maxY+g.rowHeight[y]; y++ {
		h := g.rowHeight[y]
		debugPos := currY + h/2
		debugDrawing[0][debugPos] = fmt.Sprintf("%d", y%10)
		currY += h
	}

	return mergeDrawings(&debugDrawing, &d, drawingCoord{1, 1})
}
