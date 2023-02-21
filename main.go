package main

import (
	"fmt"
	"strings"
)

const boxBorderWidth = 1
const boxBorderPadding = 1

type coord struct {
	x int
	y int
}
type drawing [][]string

func main() {
	graph := map[string][]string{"A": {"B", "C"}}

	fmt.Println("Graph: ", graph)
	d := mkDrawing(0, 0)
	for node, edges := range graph {
		nd := drawBox(node)
		d = *mergeDrawings(&d, nd, coord{0, 0})
		for i, edge := range edges {
			fmt.Println("Edge: ", edge)
			ed := drawBox(edge)
			d = *mergeDrawings(&d, ed, coord{5, 5 * i})
		}
	}
	s := drawingToString(&d)
	fmt.Println(s)
}
func drawBox(text string) *drawing {
	from := coord{0, 0}
	// -1 because we start at 0
	to := coord{len(text) + boxBorderPadding*2 + boxBorderWidth*2 - 1, boxBorderWidth*2 + boxBorderPadding*2}
	boxDrawing := mkDrawing(Max(from.x, to.x), Max(from.y, to.y))
	fmt.Println("Drawing box from ", from, " to ", to)
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

func drawArrow(from coord, to coord) *drawing {
	arrowDrawing := mkDrawing(Max(from.x, to.x), Max(from.y, to.y))
	fmt.Println("Drawing arrow from ", from, " to ", to)
	// Find the coord where the arrow should rotate
	rotateCoord := coord{from.x, to.y}
	// Draw from start to rotate
	for y := from.y; y <= rotateCoord.y; y++ {
		arrowDrawing[rotateCoord.x][y] = "|" // Vertical line
	}
	// Draw from rotate to end
	for x := rotateCoord.x; x <= to.x; x++ {
		arrowDrawing[x][rotateCoord.y] = "-" // Horizontal line
	}
	arrowDrawing[rotateCoord.x][rotateCoord.y] = "+" // Corner
	return &arrowDrawing
}
func mergeDrawings(d1 *drawing, d2 *drawing, mergeCoord coord) *drawing {
	maxX1, maxY1 := getDrawingSize(d1)
	maxX2, maxY2 := getDrawingSize(d2)
	maxX := Max(maxX1, maxX2+mergeCoord.x)
	maxY := Max(maxY1, maxY2+mergeCoord.y)
	mergedDrawing := mkDrawing(maxX, maxY)
	// Copy d1
	for x := 0; x < maxX1; x++ {
		for y := 0; y < maxY1; y++ {
			mergedDrawing[x][y] = (*d1)[x][y]
		}
	}
	// Copy d2 with offset
	for x := 0; x < maxX2; x++ {
		for y := 0; y < maxY2; y++ {
			mergedDrawing[x+mergeCoord.x][y+mergeCoord.y] = (*d2)[x][y]
		}
	}
	return &mergedDrawing
}

func drawingToString(d *drawing) string {
	maxX, maxY := getDrawingSize(d)
	s := make([]string, maxY)
	for y := 0; y < maxY; y++ {
		lineBuilder := strings.Builder{}
		for x := 0; x < maxX; x++ {
			lineBuilder.WriteString((*d)[x][y])
		}
		s[y] = lineBuilder.String()

	}
	return strings.Join(s, "\n")
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

func getDrawingSize(d *drawing) (int, int) {
	return len(*d), len((*d)[0])
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
