package main

import (
	"fmt"
	"strings"
)

const boxBorderWidth = 1
const boxBorderPadding = 1
const paddingBetweenX = 5
const paddingBetweenY = 2
const boxHeight = boxBorderPadding*2 + boxBorderWidth*2 + 1

type coord struct {
	x int
	y int
}
type drawing [][]string
type graph map[string][]string

func main() {
	totalGraph := graph{"Some text": {"B", "C"}, "B": {"D"}, "E": {"F", "G", "H"}}

	fmt.Println("Graph: ", totalGraph)
	totalDrawing := mkDrawing(0, 0)
	for node, edges := range totalGraph {
		nodeSubdrawing, _ := drawNodeWithEdges(totalGraph, node, edges)
		totalWidth, _ := getDrawingSize(&totalDrawing)
		totalDrawing = *mergeDrawings(&totalDrawing, nodeSubdrawing, coord{totalWidth, 0})
	}
	s := drawingToString(&totalDrawing)
	fmt.Println(s)
}

func drawNodeWithEdges(subGraph graph, node string, edges []string) (*drawing, graph) {
	// Remove the node from the subgraph, so we don't draw it again.
	delete(subGraph, node)

	nodeSubdrawing := mkDrawing(0, 0)
	nd := drawBox(node)
	nodeSubdrawing = *mergeDrawings(&nodeSubdrawing, nd, coord{0, 0})
	nodeWidth, nodeHeight := getDrawingSize(nd)
	for i, edge := range edges {
		fmt.Println("Edge: ", edge)
		edgeDrawing := drawBox(edge)
		edgeStart := coord{nodeWidth + paddingBetweenX, (paddingBetweenY + boxHeight) * i}
		arrowFrom := getArrowStart(nodeWidth, nodeHeight, coord{0, 0}, edgeStart)
		arrowTo := coord{edgeStart.x, edgeStart.y + boxHeight/2}
		arrowDrawing := drawArrow(arrowFrom, arrowTo)
		nodeSubdrawing = *mergeDrawings(&nodeSubdrawing, edgeDrawing, edgeStart)
		nodeSubdrawing = *mergeDrawings(&nodeSubdrawing, arrowDrawing, coord{0, 0})
		if _, ok := subGraph[edge]; ok {
			edgeSubdrawing, _ := drawNodeWithEdges(subGraph, edge, subGraph[edge])
			nodeSubdrawing = *mergeDrawings(&nodeSubdrawing, edgeSubdrawing, edgeStart)
		}
	}
	return &nodeSubdrawing, subGraph
}

func doDrawingsCollide(drawing1 *drawing, drawing2 *drawing, offset coord) bool {
	// Check if any of the drawing2 characters overlap with drawing1 characters.
	// The offset is the coord of drawing2 relative to drawing1.
	drawing1Width, drawing1Height := getDrawingSize(drawing1)
	drawing2Width, drawing2Height := getDrawingSize(drawing2)
	for x := 0; x < drawing2Width; x++ {
		for y := 0; y < drawing2Height; y++ {
			// Check if drawing2[x][y] overlaps with drawing1[x+offset.x][y+offset.y]
			if x+offset.x >= 0 && x+offset.x < drawing1Width &&
				y+offset.y >= 0 && y+offset.y < drawing1Height &&
				(*drawing2)[x][y] != " " &&
				(*drawing1)[x+offset.x][y+offset.y] != " " {
				return true
			}
		}
	}
	return false
}

func getArrowStart(startBoxWidth int, startBoxHeight int, from coord, to coord) coord {
	// Find the coord on the first box where the arrow should start.
	// This is the middle of one of the sides, depending on the direction of the arrow.
	// Note that the coord returned is relative to the start box.
	if from.x == to.x {
		// Vertical arrow
		if from.y < to.y {
			// Down
			return coord{startBoxWidth / 2, startBoxHeight}
		} else {
			// Up
			return coord{startBoxWidth / 2, 0}
		}
	} else if from.y == to.y {
		// Horizontal arrow
		if from.x < to.x {
			// Right
			return coord{startBoxWidth, startBoxHeight / 2}
		} else {
			// Left
			return coord{0, startBoxHeight / 2}
		}
	} else {
		// Diagonal arrow
		if from.x < to.x {
			// Right
			if from.y < to.y {
				// Down
				return coord{startBoxWidth / 2, startBoxHeight}
			} else {
				// Up
				return coord{startBoxWidth / 2, 0}
			}
		} else {
			// Left
			if from.y < to.y {
				// Down
				return coord{startBoxWidth / 2, startBoxHeight}
			} else {
				// Up
				return coord{startBoxWidth / 2, 0}
			}
		}
	}
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
	for x := rotateCoord.x; x < to.x; x++ {
		arrowDrawing[x][rotateCoord.y] = "-" // Horizontal line
	}
	if from.x != to.x && from.y != to.y {
		arrowDrawing[rotateCoord.x][rotateCoord.y] = "+" // Corner
	}
	arrowDrawing[to.x-1][to.y] = ">" // Arrow head
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
			c := (*d2)[x][y]
			if c != " " {
				mergedDrawing[x+mergeCoord.x][y+mergeCoord.y] = (*d2)[x][y]
			}
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
