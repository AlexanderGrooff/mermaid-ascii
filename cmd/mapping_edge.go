package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type edge struct {
	from            *node
	to              *node
	text            string
	isBidirectional bool
	path            []gridCoord
	labelLine       []gridCoord
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
			if lineWidth > fallbackLineSize {
				fallbackLineSize = lineWidth
				fallbackLine = line
			}
			continue
		}
		if lineWidth >= lenLabel {
			largestLine = line
			break
		}
		if lineWidth > largestLineSize {
			largestLineSize = lineWidth
			largestLine = line
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
	labelPadding := 3 // Wrap with -{label}-> (dashes + end arrowhead, 3 char)
	if e.isBidirectional {
		labelPadding = 4 // Wrap with <-{label}-> (start arrowhead+ dashes + end arrowhead, 4 char)
	}
	log.Debugf("Increasing column width for column %v from size %v to %v", middleX, g.columnWidth[middleX], lenLabel+labelPadding)
	g.columnWidth[middleX] = Max(g.columnWidth[middleX], lenLabel+labelPadding)
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
