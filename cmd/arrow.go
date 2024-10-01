package cmd

import (
	log "github.com/sirupsen/logrus"
)

func (g *graph) getPath(from gridCoord, to gridCoord, prevSteps []gridCoord) []gridCoord {
	// Figure out what path the arrow should take by traversing the grid recursively.
	var nextPos gridCoord
	log.Debugf("Looking for path from %v to %v", from, to)

	if from == to {
		return []gridCoord{}
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
		panic("Out of bounds")
	}
	if (absX == 0 && absY == 1) || (absX == 1 && absY == 0) {
		// Can go directly to the target
		return []gridCoord{to}
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

	item := nextPos
	slice := g.getPath(nextPos, to, append(prevSteps, nextPos))
	return append([]gridCoord{item}, slice...)
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
	// TODO: How to determine where the arrow should start/end?
	path := g.getPath(from, to, []gridCoord{})
	linesDrawn := g.drawPath(from, to, path)
	g.drawArrowHead(linesDrawn[len(linesDrawn)-1])
	// TODO: draw label. Maybe on a step that has long enough X? Or the longest X? How do you set column width? Maybe determine gridCoord for label?
}

func (g *graph) drawPath(from gridCoord, to gridCoord, path []gridCoord) [][]drawingCoord {
	log.Debugf("Drawing arrow from %v to %v with path %v", from, to, path)

	d := g.drawing
	previousCoord := from
	linesDrawn := make([][]drawingCoord, 0)
	var previousDrawingCoord drawingCoord
	// var dir direction
	// var oppositeDir direction
	for idx, nextCoord := range path {
		// if idx == 0 {
		// 	// Only the first/last step goes from/to the edges of a node. Intermediate steps cross the middle.
		// 	dir = determineDirection(genericCoord(previousCoord), genericCoord(nextCoord))
		// } else {
		// 	dir = Middle
		// }
		// if idx == len(path)-1 {
		// 	oppositeDir = determineDirection(genericCoord(previousCoord), genericCoord(nextCoord)).getOpposite()
		// } else {
		// 	oppositeDir = Middle
		// }
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
	return linesDrawn
}

func (g *graph) drawArrowHead(line []drawingCoord) {
	// Determine the direction of the arrow for the last step
	from := line[0]
	lastPos := line[len(line)-1]
	dir := determineDirection(genericCoord(from), genericCoord(lastPos))
	switch dir {
	case Up:
		(*g.drawing)[lastPos.x][lastPos.y] = "^"
	case Down:
		(*g.drawing)[lastPos.x][lastPos.y] = "v"
	case Left:
		(*g.drawing)[lastPos.x][lastPos.y] = "<"
	case Right:
		(*g.drawing)[lastPos.x][lastPos.y] = ">"
	case UpperRight:
		(*g.drawing)[lastPos.x][lastPos.y] = "┐"
	case UpperLeft:
		(*g.drawing)[lastPos.x][lastPos.y] = "┌"
	case LowerRight:
		(*g.drawing)[lastPos.x][lastPos.y] = "┘"
	case LowerLeft:
		(*g.drawing)[lastPos.x][lastPos.y] = "└"
	default:
		(*g.drawing)[lastPos.x][lastPos.y] = "+"
	}
}
