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
	startDir  direction
	endDir    direction
}

func (g *graph) determinePath(e *edge) {
	// Get both paths and use least amount of steps
	var preferredPath, alternativePath []gridCoord
	var err error
	preferredDir, preferredOppositeDir, alternativeDir, alternativeOppositeDir := determineStartAndEndDir(e)
	from := e.from.gridCoord.Direction(preferredDir)
	to := e.to.gridCoord.Direction(preferredOppositeDir)
	log.Debugf("Determining preferred path from %v (direction %v) to %v (direction %v)", *e.from, preferredDir, *e.to, preferredOppositeDir)

	// Get preferred path
	preferredPath, err = g.getPath(from, to)
	// preferredPath = append([]gridCoord{from}, preferredPath...) // TODO: how to do add 'from' to path nicely?
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
	// alternativePath = append([]gridCoord{from}, alternativePath...)
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
