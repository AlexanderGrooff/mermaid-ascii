package graph

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

// edgeStroke is the line style drawn along an edge's path. Only ASCII
// output (g.useAscii) currently distinguishes them; Unicode output always
// renders a solid line regardless of stroke.
type edgeStroke int

const (
	strokeSolid edgeStroke = iota
	strokeDotted
	strokeThick
)

// edgeHead is the glyph drawn at an edge's arrowhead end. headNone means no
// glyph is drawn at all (an open link like ---). Only ASCII output currently
// distinguishes them; Unicode output always draws a directional arrowhead.
type edgeHead int

const (
	headArrow edgeHead = iota
	headCircle
	headCross
	headNone
)

type edge struct {
	from            *node
	to              *node
	text            string
	isBidirectional bool
	stroke          edgeStroke
	head            edgeHead
	path            []gridCoord
	labelLine       []gridCoord
	fanout          bool
	startDir        direction
	endDir          direction
}

func (g *graph) determinePath(e *edge) {
	key := newEdgePair(e.from.index, e.to.index)
	duplicateIndex := g.edgeCounts[key]

	if startDir, endDir, ok := g.parallelDirections(e, duplicateIndex); ok {
		from := e.from.gridCoord.Direction(startDir)
		to := e.to.gridCoord.Direction(endDir)
		if path, err := g.getPath(from, to); err == nil {
			e.startDir = startDir
			e.endDir = endDir
			e.path = mergePath(path)
			g.edgeCounts[key]++
			return
		}
	}

	// Get both paths and use least amount of steps
	var preferredPath, alternativePath []gridCoord
	var from, to gridCoord
	var err error
	preferredDir, preferredOppositeDir, alternativeDir, alternativeOppositeDir := g.determineStartAndEndDir(e)

	from = e.from.gridCoord.Direction(preferredDir)
	to = e.to.gridCoord.Direction(preferredOppositeDir)
	log.Debugf("Determining preferred path from %v (direction %v) to %v (direction %v)", *e.from, preferredDir, *e.to, preferredOppositeDir)

	// Get preferred path
	preferredPath, err = g.getPath(from, to)
	if err != nil {
		fmt.Printf("Error getting path from %v to %v: %v", from, to, err)
		// This is a big assumption, but if we can't get the preferred path, we assume the alternative path is better
		e.startDir = alternativeDir
		e.endDir = alternativeOppositeDir
		e.path = alternativePath
		return
	}
	preferredPath = mergePath(preferredPath)

	// Alternative path
	from = e.from.gridCoord.Direction(alternativeDir)
	to = e.to.gridCoord.Direction(alternativeOppositeDir)
	log.Debugf("Determining alternative path from %v (direction %v) to %v (direction %v)", *e.from, alternativeDir, *e.to, alternativeOppositeDir)

	alternativePath, err = g.getPath(from, to)
	if err != nil {
		fmt.Printf("Error getting path from %v to %v: %v", from, to, err)
		e.startDir = preferredDir
		e.endDir = preferredOppositeDir
		e.path = preferredPath
	}
	alternativePath = mergePath(alternativePath)

	nrStepsPreferred := len(preferredPath)
	nrStepsAlternative := len(alternativePath)
	if nrStepsPreferred <= nrStepsAlternative {
		log.Debugf("Using preferred path with %v steps instead of alternative path with %v steps", nrStepsPreferred, nrStepsAlternative)
		e.startDir = preferredDir
		e.endDir = preferredOppositeDir
		e.path = preferredPath
	} else {
		log.Debugf("Using alternative path with %v steps instead of alternative path with %v steps", nrStepsAlternative, nrStepsPreferred)
		e.startDir = alternativeDir
		e.endDir = alternativeOppositeDir
		e.path = alternativePath
	}
	g.edgeCounts[key]++
}

func (g *graph) parallelDirections(e *edge, duplicateIndex int) (direction, direction, bool) {
	if duplicateIndex == 0 {
		return Middle, Middle, false
	}

	dir := determineDirection(genericCoord(*e.from.gridCoord), genericCoord(*e.to.gridCoord))
	switch {
	case g.graphDirection == "LR" && (dir == Right || dir == Left):
		options := [][2]direction{{Down, Down}, {Up, Up}}
		if duplicateIndex-1 < len(options) {
			return options[duplicateIndex-1][0], options[duplicateIndex-1][1], true
		}
	case g.graphDirection == "TD" && (dir == Down || dir == Up):
		options := [][2]direction{{Right, Right}, {Left, Left}}
		if duplicateIndex-1 < len(options) {
			return options[duplicateIndex-1][0], options[duplicateIndex-1][1], true
		}
	}

	return Middle, Middle, false
}

func (g *graph) determineLabelLine(e *edge) {
	// What line on the path should the label be placed?
	lenLabel := len(e.text)
	if lenLabel == 0 {
		return
	}
	// Widening a column that is occupied by a node would push that node's
	// border out, leaving a visible gap between the box and any incoming
	// arrowhead. Prefer label-line candidates whose target column is a free
	// edge corridor; only fall back to a node column if no corridor segment
	// is available.
	prevStep := e.path[0]
	var largestLine []gridCoord
	var largestLineSize int
	var fallbackLine []gridCoord
	var fallbackLineSize int
	for _, step := range e.path[1:] {
		line := []gridCoord{gridCoord(prevStep), gridCoord(step)}
		prevStep = step
		lineWidth := g.calculateLineWidth(line)
		if g.isNodeColumn(labelMiddleX(line)) {
			if g.labelBoundsAvailable(e, line) && lineWidth > fallbackLineSize {
				fallbackLineSize = lineWidth
				fallbackLine = line
			}
			continue
		}
		if g.labelLineAvailable(e, line) && g.labelBoundsAvailable(e, line) {
			if lineWidth >= lenLabel {
				largestLine = line
				break
			}
			if lineWidth > largestLineSize {
				largestLineSize = lineWidth
				largestLine = line
			}
		}
	}
	if largestLine == nil {
		// A fan-out can share the first corridor while offering a later
		// corridor wide enough for a label. Look for a short, unused section
		// centered on a free column rather than overwriting another label or
		// widening a node column.
		for _, line := range g.unusedLabelSubsegments(e) {
			lineWidth := g.calculateLineWidth(line)
			if lineWidth >= lenLabel {
				largestLine = line
				break
			}
			if lineWidth > largestLineSize {
				largestLineSize = lineWidth
				largestLine = line
			}
		}
	}
	if largestLine == nil {
		largestLine = fallbackLine
	}
	if largestLine == nil {
		// Path only had a single segment that lives on a node column; use it
		// rather than dropping the label entirely.
		largestLine = []gridCoord{e.path[0], e.path[1]}
	}

	middleX := labelMiddleX(largestLine)
	labelPadding := labelPaddingForEdge(e)
	log.Debugf("Increasing column width for column %v from size %v to %v", middleX, g.columnWidth[middleX], lenLabel+labelPadding)
	g.columnWidth[middleX] = Max(g.columnWidth[middleX], lenLabel+labelPadding)
	g.ensureLabelLineClear(e, largestLine)
	log.Debugf("New column sizes: %v", g.columnWidth)
	e.labelLine = largestLine
}

func labelMiddleX(line []gridCoord) int {
	minX, maxX := line[0].x, line[1].x
	if minX > maxX {
		minX, maxX = maxX, minX
	}
	return minX + (maxX-minX)/2
}

func (g *graph) labelLineAvailable(edge *edge, line []gridCoord) bool {
	for _, other := range g.edges {
		if other == edge || len(other.labelLine) == 0 {
			continue
		}
		if labelsOverlap(g, edge, line, other) {
			return false
		}
	}
	return true
}

func labelPaddingForEdge(e *edge) int {
	if e.isBidirectional {
		return 4 // start arrowhead + dashes + end arrowhead
	}
	return 3 // dashes + end arrowhead
}

// labelBoundsAvailable keeps the label layer from overwriting drawing elements
// that are not redrawn after labels are merged. It is intentionally based on
// coordinates rather than glyphs so it also protects non-ASCII arrowheads.
func (g *graph) labelBoundsAvailable(e *edge, line []gridCoord) bool {
	drawingLine := g.lineToDrawing(line)
	if e.isBidirectional {
		drawingLine = insetLine(drawingLine, 2, 2)
	} else {
		drawingLine = insetLine(drawingLine, 1, 2)
	}
	start, end, y := labelBounds(drawingLine, e.text)
	for _, other := range g.edges {
		if len(other.path) < 2 || other.head == headNone {
			continue
		}
		arrow := g.arrowheadPosition(other)
		if arrow.y == y && rangesOverlap(start, end, arrow.x, arrow.x) {
			return false
		}
		if other.isBidirectional {
			arrow = g.arrowheadPositionFromStart(other)
			if arrow.y == y && rangesOverlap(start, end, arrow.x, arrow.x) {
				return false
			}
		}
	}
	for _, n := range g.nodes {
		if g.labelOverlapsNodeBorder(n, start, end, y) {
			return false
		}
	}
	return true
}

func (g *graph) arrowheadPosition(e *edge) drawingCoord {
	last := g.gridToDrawingCoord(e.path[len(e.path)-1], nil)
	if len(e.path) < 2 {
		return last
	}
	previous := g.gridToDrawingCoord(e.path[len(e.path)-2], nil)
	dir := determineDirection(genericCoord(previous), genericCoord(last))
	dx, dy := drawingStep(dir)
	return drawingCoord{x: last.x - dx, y: last.y - dy}
}

func (g *graph) arrowheadPositionFromStart(e *edge) drawingCoord {
	first := g.gridToDrawingCoord(e.path[0], nil)
	if len(e.path) < 2 {
		return first
	}
	next := g.gridToDrawingCoord(e.path[1], nil)
	dir := determineDirection(genericCoord(next), genericCoord(first))
	dx, dy := drawingStep(dir)
	return drawingCoord{x: first.x - dx, y: first.y - dy}
}

func drawingStep(dir direction) (int, int) {
	switch dir {
	case Up:
		return 0, -1
	case Down:
		return 0, 1
	case Left:
		return -1, 0
	case Right:
		return 1, 0
	}
	return 0, 0
}

func (g *graph) labelOverlapsNodeBorder(n *node, start, end, y int) bool {
	if n.gridCoord == nil {
		return false
	}
	from := g.gridToDrawingCoord(*n.gridCoord, nil)
	to := drawingCoord{
		x: from.x + g.columnWidth[n.gridCoord.x] + g.columnWidth[n.gridCoord.x+1],
		y: from.y + g.rowHeight[n.gridCoord.y] + g.rowHeight[n.gridCoord.y+1],
	}
	if y != from.y && y != to.y {
		return false
	}
	return rangesOverlap(start, end, from.x, to.x)
}

// ensureLabelLineClear relocates a label corridor only when its rendered
// bounds would touch a node border or arrowhead. Increasing the gap's grid
// dimension preserves the existing path and leaves normal-spacing layouts
// unchanged.
func (g *graph) ensureLabelLineClear(e *edge, line []gridCoord) {
	for attempt := 0; attempt < 8 && !g.labelBoundsAvailable(e, line); attempt++ {
		dir := determineDirection(genericCoord(line[0]), genericCoord(line[1]))
		switch dir {
		case Down:
			g.rowHeight[line[0].y+1] += 2
		case Up:
			g.rowHeight[line[0].y-1] += 2
		case Right:
			g.columnWidth[line[0].x+1] += 2
		case Left:
			g.columnWidth[line[0].x-1] += 2
		default:
			return
		}
	}
}

func labelsOverlap(g *graph, firstEdge *edge, firstLine []gridCoord, secondEdge *edge) bool {
	firstDrawingLine := g.lineToDrawing(firstLine)
	secondDrawingLine := g.lineToDrawing(secondEdge.labelLine)
	if firstEdge.isBidirectional {
		firstDrawingLine = insetLine(firstDrawingLine, 2, 2)
	} else {
		firstDrawingLine = insetLine(firstDrawingLine, 1, 2)
	}
	if secondEdge.isBidirectional {
		secondDrawingLine = insetLine(secondDrawingLine, 2, 2)
	} else {
		secondDrawingLine = insetLine(secondDrawingLine, 1, 2)
	}
	firstStart, firstEnd, firstY := labelBounds(firstDrawingLine, firstEdge.text)
	secondStart, secondEnd, secondY := labelBounds(secondDrawingLine, secondEdge.text)
	return firstY == secondY && rangesOverlap(firstStart, firstEnd, secondStart, secondEnd)
}

func labelBounds(line []drawingCoord, label string) (int, int, int) {
	middleX := line[0].x + (line[1].x-line[0].x)/2
	middleY := line[0].y + (line[1].y-line[0].y)/2
	start := middleX - len(label)/2
	return start, start + len(label) - 1, middleY
}

func rangesOverlap(firstStart, firstEnd, secondStart, secondEnd int) bool {
	firstMin, firstMax := firstStart, firstEnd
	if firstMin > firstMax {
		firstMin, firstMax = firstMax, firstMin
	}
	secondMin, secondMax := secondStart, secondEnd
	if secondMin > secondMax {
		secondMin, secondMax = secondMax, secondMin
	}
	return firstMin <= secondMax && secondMin <= firstMax
}

func (g *graph) unusedLabelSubsegments(edge *edge) [][]gridCoord {
	var segments [][]gridCoord
	for i, start := range edge.path[:len(edge.path)-1] {
		end := edge.path[i+1]
		if start.y != end.y || start.x == end.x {
			continue
		}
		minX, maxX := start.x, end.x
		if minX > maxX {
			minX, maxX = maxX, minX
		}
		for middleX := minX + 1; middleX < maxX; middleX++ {
			if g.isNodeColumn(middleX) {
				continue
			}
			line := []gridCoord{{x: middleX - 1, y: start.y}, {x: middleX + 1, y: start.y}}
			if g.labelLineAvailable(edge, line) && g.labelBoundsAvailable(edge, line) {
				segments = append(segments, line)
			}
		}
	}
	return segments
}

// isNodeColumn reports whether grid column x is occupied by any node.
// Widening such a column distorts the box that owns it.
func (g *graph) isNodeColumn(x int) bool {
	for _, n := range g.nodes {
		if n.gridCoord == nil {
			continue
		}
		if x >= n.gridCoord.x && x <= n.gridCoord.x+2 {
			return true
		}
	}
	return false
}

func (g graph) calculateLineWidth(line []gridCoord) int {
	totalSize := 0
	for _, c := range line {
		totalSize += g.columnWidth[c.x]
	}
	return totalSize
}
