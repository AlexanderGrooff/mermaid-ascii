package cmd

import (
	"fmt"
	"slices"

	log "github.com/sirupsen/logrus"
)

func (g *graph) getPath(from gridCoord, to gridCoord, prevSteps []gridCoord) ([]gridCoord, error) {
	// Figure out what path the arrow should take by traversing the grid recursively.
	var nextPos gridCoord
	log.Debugf("Looking for path from %v to %v", from, to)

	if from == to {
		return []gridCoord{}, nil
	}

	deltaX := to.x - from.x
	deltaY := to.y - from.y

	// These variables are used to determine what direction to head towards first. If the target is above (e.g. deltaY < 0)
	// then we want to go up first.
	var preferredXDirection int
	var preferredYDirection int
	if deltaX > 0 {
		preferredXDirection = 1
	} else {
		preferredXDirection = -1
	}
	if deltaY > 0 {
		preferredYDirection = 1
	} else {
		preferredYDirection = -1
	}

	absX := Abs(deltaX)
	absY := Abs(deltaY)
	// TODO: make nicer
	if absX > 50 || absY > 50 {
		return []gridCoord{}, fmt.Errorf("Out of bounds")
	}
	if (absX == 0 && absY == 1) || (absX == 1 && absY == 0) {
		// Can go directly to the target
		return []gridCoord{to}, nil
	} else if deltaY == 0 {
		if deltaX > 0 && g.isFreeInGrid(gridCoord{x: from.x + 1, y: from.y}) && !hasStepBeenTaken(gridCoord{x: from.x + 1, y: from.y}, prevSteps) {
			nextPos = gridCoord{x: from.x + 1, y: from.y}
		} else if g.isFreeInGrid(gridCoord{x: from.x - 1, y: from.y}) && !hasStepBeenTaken(gridCoord{x: from.x - 1, y: from.y}, prevSteps) {
			nextPos = gridCoord{x: from.x - 1, y: from.y}
		} else {
			// We're on the correct Y level, but the X direction is not free. We have to go around either from
			// above or below our neighbour.
			// TODO: prevent from taking steps that have already been assigned
			// TODO: diagonal?
			// We start by trying below first
			if g.isFreeInGrid(gridCoord{x: from.x, y: from.y + 1}) && !hasStepBeenTaken(gridCoord{x: from.x, y: from.y + 1}, prevSteps) {
				nextPos = gridCoord{x: from.x, y: from.y + 1}
			} else {
				// TODO: what if all positions are taken?
				nextPos = gridCoord{x: from.x, y: from.y - 1}
			}
		}
	} else if deltaX == 0 {
		if deltaY > 0 && g.isFreeInGrid(gridCoord{x: from.x, y: from.y + 1}) && !hasStepBeenTaken(gridCoord{x: from.x, y: from.y + 1}, prevSteps) {
			nextPos = gridCoord{x: from.x, y: from.y + 1}
		} else if g.isFreeInGrid(gridCoord{x: from.x, y: from.y - 1}) && !hasStepBeenTaken(gridCoord{x: from.x, y: from.y - 1}, prevSteps) {
			nextPos = gridCoord{x: from.x, y: from.y - 1}
		} else {
			// We're on the correct X level, but the Y direction is not free. We have to go around either from
			// above or below our neighbour.
			// TODO: prevent from taking steps that have already been assigned
			// TODO: diagonal?
			// We start by trying right first
			if g.isFreeInGrid(gridCoord{x: from.x + 1, y: from.y}) && !hasStepBeenTaken(gridCoord{x: from.x + 1, y: from.y}, prevSteps) {
				nextPos = gridCoord{x: from.x + 1, y: from.y}
			} else {
				// TODO: what if all positions are taken?
				nextPos = gridCoord{x: from.x - 1, y: from.y}
			}
		}
	} else {
		// If we're in LR, we first go vertical then horizontal.
		if graphDirection == "LR" {
			if g.isFreeInGrid(gridCoord{x: from.x, y: from.y + preferredYDirection}) && !hasStepBeenTaken(gridCoord{x: from.x, y: from.y + preferredYDirection}, prevSteps) {
				nextPos = gridCoord{x: from.x, y: from.y + preferredYDirection}
			} else if g.isFreeInGrid(gridCoord{x: from.x + preferredXDirection, y: from.y}) && !hasStepBeenTaken(gridCoord{x: from.x + preferredXDirection, y: from.y}, prevSteps) {
				// Vertical is blocked, let's try horizontal
				nextPos = gridCoord{x: from.x + preferredXDirection, y: from.y}
			} else if g.isFreeInGrid(gridCoord{x: from.x, y: from.y - preferredYDirection}) && !hasStepBeenTaken(gridCoord{x: from.x, y: from.y - preferredYDirection}, prevSteps) {
				nextPos = gridCoord{x: from.x, y: from.y - preferredYDirection}
			} else {
				// TODO: Diagonal?
				// TODO: what about inbetween nodes, on half/grid coords?
				nextPos = gridCoord{x: from.x - preferredXDirection, y: from.y}
			}
		} else if graphDirection == "TD" {
			// If we're in TD, we first go horizontal then vertical.
			if g.isFreeInGrid(gridCoord{x: from.x + preferredXDirection, y: from.y}) && !hasStepBeenTaken(gridCoord{x: from.x + preferredXDirection, y: from.y}, prevSteps) {
				nextPos = gridCoord{x: from.x + preferredXDirection, y: from.y}
			} else if g.isFreeInGrid(gridCoord{x: from.x, y: from.y + preferredYDirection}) && !hasStepBeenTaken(gridCoord{x: from.x, y: from.y + preferredYDirection}, prevSteps) {
				// Horizontal is blocked, let's try vertical
				nextPos = gridCoord{x: from.x, y: from.y + preferredYDirection}
			} else if g.isFreeInGrid(gridCoord{x: from.x - preferredXDirection, y: from.y}) && !hasStepBeenTaken(gridCoord{x: from.x - preferredXDirection, y: from.y}, prevSteps) {
				nextPos = gridCoord{x: from.x - preferredXDirection, y: from.y}
			} else {
				// TODO: Diagonal?
				nextPos = gridCoord{x: from.x, y: from.y - preferredYDirection}
			}
		}
	}

	currSteps := append(prevSteps, nextPos)
	slice, err := g.getPath(nextPos, to, currSteps)
	if err != nil {
		return currSteps, err
	}
	return append([]gridCoord{nextPos}, slice...), nil
}

func (g *graph) isFreeInGrid(c gridCoord) bool {
	return g.grid[c] == nil
}

func hasStepBeenTaken(step gridCoord, steps []gridCoord) bool {
	for _, s := range steps {
		if s == step {
			log.Debugf("Step %v has been taken", s)
			return true
		}
	}
	return false
}

func (g *graph) drawArrow(from gridCoord, to gridCoord, label string) {
	path, err := g.getPath(from, to, []gridCoord{from})
	path = append([]gridCoord{from}, path...) // TODO: how to do this nicely in getPath?
	if err != nil {
		fmt.Printf("Error getting path from %v to %v: %v", from, to, err)
	}
	path = mergePath(path)
	log.Debugf("Drawing arrow from %v to %v with path %v", from, to, path)
	dLabel := g.drawArrowLabel(path, label)
	dPath, linesDrawn := g.drawPath(path)
	dHead := g.drawArrowHead(linesDrawn[len(linesDrawn)-1])
	g.drawing = mergeDrawings(g.drawing, dPath, drawingCoord{0, 0})
	g.drawing = mergeDrawings(g.drawing, dHead, drawingCoord{0, 0})
	g.drawing = mergeDrawings(g.drawing, dLabel, drawingCoord{0, 0})
}

func mergePath(path []gridCoord) []gridCoord {
	// If two steps are in the same direction, merge them to one step.
	if len(path) <= 2 {
		return path
	}
	indexToRemove := []int{}
	step0 := path[0]
	step1 := path[1]
	for idx, step2 := range path[2:] {
		prevDir := determineDirection(genericCoord(step0), genericCoord(step1))
		dir := determineDirection(genericCoord(step1), genericCoord(step2))
		if prevDir == dir {
			log.Debugf("Removing %v from path", step1)
			indexToRemove = append(indexToRemove, idx+1) // +1 because we skip the initial step
		}
		step0 = step1
		step1 = step2
	}
	newPath := []gridCoord{}
	for idx, step := range path {
		if !slices.Contains(indexToRemove, idx) {
			newPath = append(newPath, step)
		}
	}
	return newPath
}

func (g *graph) drawPath(path []gridCoord) (*drawing, [][]drawingCoord) {
	d := copyCanvas(g.drawing)
	previousCoord := path[0]
	linesDrawn := make([][]drawingCoord, 0)
	var previousDrawingCoord drawingCoord
	for idx, nextCoord := range path[1:] {
		previousDrawingCoord = g.gridToDrawingCoord(previousCoord, nil)
		nextDrawingCoord := g.gridToDrawingCoord(nextCoord, nil)
		if previousDrawingCoord.Equals(nextDrawingCoord) {
			log.Debugf("Skipping drawing identical line on %v", nextCoord)
			continue
		}
		if idx == 0 {
			// Don't cross the node border
			linesDrawn = append(linesDrawn, d.drawLine(previousDrawingCoord, nextDrawingCoord, 1, -1))
		} else {
			linesDrawn = append(linesDrawn, d.drawLine(previousDrawingCoord, nextDrawingCoord, 0, -1))
		}
		previousCoord = nextCoord
	}
	return d, linesDrawn
}

func (g *graph) drawArrowHead(line []drawingCoord) *drawing {
	d := *(copyCanvas(g.drawing))
	// Determine the direction of the arrow for the last step
	from := line[0]
	lastPos := line[len(line)-1]
	dir := determineDirection(genericCoord(from), genericCoord(lastPos))
	switch dir {
	case Up:
		d[lastPos.x][lastPos.y] = "^"
	case Down:
		d[lastPos.x][lastPos.y] = "v"
	case Left:
		d[lastPos.x][lastPos.y] = "<"
	case Right:
		d[lastPos.x][lastPos.y] = ">"
	case UpperRight:
		d[lastPos.x][lastPos.y] = "┐"
	case UpperLeft:
		d[lastPos.x][lastPos.y] = "┌"
	case LowerRight:
		d[lastPos.x][lastPos.y] = "┘"
	case LowerLeft:
		d[lastPos.x][lastPos.y] = "└"
	default:
		d[lastPos.x][lastPos.y] = "+"
	}
	return &d
}

func (g *graph) drawArrowLabel(path []gridCoord, label string) *drawing {
	d := copyCanvas(g.drawing)
	lenLabel := len(label)
	if lenLabel == 0 {
		return d
	}
	prevStep := g.gridToDrawingCoord(path[0], nil)
	var gridX, maxX, minX, largestLineSize int
	// Init to first line if we find nothing else
	largestLine := []drawingCoord{prevStep, g.gridToDrawingCoord(path[1], nil)}
	largestLineSize = 0
	for _, s := range path[1:] {
		step := g.gridToDrawingCoord(s, nil)
		if step.x > prevStep.x {
			minX = prevStep.x
			maxX = step.x
		} else {
			minX = step.x
			maxX = step.x
		}
		if (maxX - minX) >= lenLabel {
			gridX = s.x
			largestLine = []drawingCoord{prevStep, step}
			break
		} else if (maxX - minX) > largestLineSize {
			largestLineSize = maxX - minX
			largestLine = []drawingCoord{prevStep, step}
		}
		prevStep = step
	}
	// TODO: this happens too late. We should not calculate drawingCoords before we do this.
	g.columnWidth[gridX] = Max(g.columnWidth[gridX], lenLabel+2)
	d.drawTextOnLine(largestLine, label)
	return d
}

func (d *drawing) drawTextOnLine(line []drawingCoord, label string) {
	// Write text in middle of the line
	//  123456789
	// |---------|
	//     123
	var minX, maxX, minY, maxY int
	if line[0].x > line[1].x {
		minX = line[1].x
		maxX = line[0].x
	} else {
		minX = line[0].x
		maxX = line[1].x
	}
	if line[0].y > line[1].y {
		minY = line[1].y
		maxY = line[0].y
	} else {
		minY = line[0].y
		maxY = line[1].y
	}
	middleX := minX + (maxX-minX)/2
	middleY := minY + (maxY-minY)/2
	startLabelCoord := drawingCoord{x: middleX - len(label)/2, y: middleY}
	d.drawText(startLabelCoord, label)
}
