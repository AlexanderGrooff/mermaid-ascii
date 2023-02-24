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
type graphData map[string][]string

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

func (g graph) getEdgesToNode(node node) []edge {
	edges := []edge{}
	for _, edge := range g.edges {
		if (&edge.to) == (&node) {
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

func (g graph) getRootNodes() []node {
	nodes := []node{}
	for _, node := range g.nodes {
		edges := g.getEdgesToNode(node)
		if len(edges) == 0 {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func (g *graph) getOrCreateRootNode(name string) node {
	// Check if the node already exists.
	for _, existingRootNode := range g.nodes {
		if existingRootNode.name == name {
			fmt.Println("Found existing root node", existingRootNode.name, "at", existingRootNode.coord)
			return existingRootNode
		}
	}
	parentCoord := g.positionNextRootNode()
	fmt.Println("Creating new root node", name, "at", parentCoord)
	parentNode := node{name: name, drawing: drawBox(name), coord: parentCoord}
	g.drawNode(parentNode)
	g.appendNode(parentNode)
	return parentNode
}

func (g graph) positionNextRootNode() coord {
	previousRootNodes := g.getRootNodes()
	return coord{len(previousRootNodes) * 10, 0}
}

func (g *graph) getOrCreateChildNode(parent node, name string) node {
	// Check if the node already exists.
	for _, existingChildNode := range g.nodes {
		if existingChildNode.name == name {
			fmt.Println("Found existing child node", existingChildNode.name, "at", existingChildNode.coord)
			return existingChildNode
		}
	}
	childNode := node{name: name, drawing: drawBox(name)}
	childCoord := g.findPositionChildNode(parent, childNode)
	childNode.setCoord(childCoord)
	g.drawNode(childNode)
	g.appendNode(childNode)
	fmt.Println("Placed child node: ", childNode.coord)
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
		fmt.Println("Placing child node", child.name, "next to parent node", parent.name)
		return coordNextToParent
	} else {
		// The child node can't be placed next to the parent node.
		// Find the last child node, and place the node under that one.
		// If there are no child nodes, place it under the parent node.
		children := g.getChildren(parent)
		if len(children) == 0 {
			fmt.Println("Couldn't find position for child node", child.name, "for parent node", parent.name)
			return coord{x: 15, y: 15}
		}
		lastChildNode := children[len(children)-1]
		_, lastChildNodeHeight := getDrawingSize(lastChildNode.drawing)
		fmt.Println("Placing child node", child.name, "under last child node", lastChildNode.name, "for parent node", parent.name)
		return coord{x: lastChildNode.coord.x, y: lastChildNode.coord.y + lastChildNodeHeight + paddingBetweenY}
	}
}

func (g graph) dimensions() (int, int) {
	return getDrawingSize(g.drawing)
}

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

func main() {
	// data := graphData{"Some text": {"B", "C"}, "B": {"D"}, "E": {"F", "G", "H"}, "C": {"D"}}
	data := graphData{"A": {"C", "D"}, "B": {"C", "D"}, "C": {"D"}}
	totalGraph := mkGraph(data)
	s := drawingToString(totalGraph.drawing)
	fmt.Println(s)
}

func mkGraph(data graphData) graph {
	g := graph{drawing: mkDrawing(0, 0)}
	for nodeName, children := range data {
		parentNode := g.getOrCreateRootNode(nodeName)
		for _, childNodeName := range children {
			childNode := g.getOrCreateChildNode(parentNode, childNodeName)
			e := edge{from: parentNode, to: childNode, text: ""}
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

func drawBox(text string) drawing {
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

	return boxDrawing
}

func drawArrow(from coord, to coord) drawing {
	// Stop arrow one character before the end coord to stop just before the target
	arrowDrawing := mkDrawing(Max(from.x, to.x), Max(from.y, to.y))
	fmt.Println("Drawing arrow from", from, "to", to)
	// Find the coord where the arrow should rotate
	rotateCoord := coord{from.x, to.y}
	// Draw from start to rotate
	if from.y <= rotateCoord.y {
		// Up
		for y := from.y; y < rotateCoord.y; y++ {
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
		for x := rotateCoord.x; x < to.x; x++ {
			arrowDrawing[x][rotateCoord.y] = "-" // Horizontal line
		}
	} else {
		for x := to.x; x < rotateCoord.x; x++ {
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
		arrowDrawing[to.x][to.y] = "<"
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
	for x := 0; x < maxX1; x++ {
		for y := 0; y < maxY1; y++ {
			mergedDrawing[x][y] = d1[x][y]
		}
	}
	// Copy d2 with offset
	for x := 0; x < maxX2; x++ {
		for y := 0; y < maxY2; y++ {
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
	s := make([]string, maxY)
	for y := 0; y < maxY; y++ {
		lineBuilder := strings.Builder{}
		for x := 0; x < maxX; x++ {
			lineBuilder.WriteString(d[x][y])
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

func getDrawingSize(d drawing) (int, int) {
	return len(d), len(d[0])
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
