package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type edge struct {
	from      *node
	to        *node
	text      string
	path      []gridCoord
	labelLine []gridCoord
}

func (g *graph) determinePath(e *edge) {
	dir, oppositeDir := determineStartAndEndDir(e)
	from := e.from.gridCoord.Direction(dir)
	to := e.to.gridCoord.Direction(oppositeDir)
	log.Debugf("Determining path from %v (direction %v) to %v (direction %v)", *e.from, dir, *e.to, oppositeDir)

	path, err := g.getPath(from, to, []gridCoord{from})
	path = append([]gridCoord{from}, path...) // TODO: how to do add 'from' to path nicely?
	if err != nil {
		fmt.Printf("Error getting path from %v to %v: %v", from, to, err)
	}
	path = mergePath(path)
	e.path = path
}

func (g *graph) determineLabelLine(e *edge) {
	// What line on the path should the label be placed?
	lenLabel := len(e.text)
	if lenLabel == 0 {
		return
	}
	prevStep := e.path[0]
	var largestLineSize int
	// Init to first line if we find nothing else
	largestLine := []gridCoord{prevStep, e.path[1]}
	largestLineSize = 0
	for _, step := range e.path[1:] {
		line := []gridCoord{gridCoord(prevStep), gridCoord(step)}
		lineWidth := g.calculateLineWidth(line)
		if lineWidth >= lenLabel {
			largestLine = line
			break
		} else if lineWidth > largestLineSize {
			largestLineSize = lineWidth
			largestLine = line
		}
		prevStep = step
	}

	var maxX, minX int
	if largestLine[0].x > largestLine[1].x {
		maxX = largestLine[0].x
		minX = largestLine[1].x
	} else {
		maxX = largestLine[1].x
		minX = largestLine[0].x
	}
	middleX := minX + (maxX-minX)/2
	log.Debugf("Increasing column width for column %v from size %v to %v", middleX, g.columnWidth[middleX], lenLabel+2)
	g.columnWidth[middleX] = Max(g.columnWidth[middleX], lenLabel+2) // Wrap with dashes + arrowhead
	log.Debugf("New column sizes: %v", g.columnWidth)
	e.labelLine = largestLine
}

func (g graph) calculateLineWidth(line []gridCoord) int {
	totalSize := 0
	for _, c := range line {
		totalSize += g.columnWidth[c.x]
	}
	return totalSize
}
