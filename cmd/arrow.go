package cmd

import (
	log "github.com/sirupsen/logrus"
)

type direction string

const (
	Up         = "Up"
	Down       = "Down"
	Left       = "Left"
	Right      = "Right"
	UpperRight = "UpperRight"
	UpperLeft  = "UpperLeft"
	LowerRight = "LowerRight"
	LowerLeft  = "LowerLeft"
)

func (d direction) getOpposite() direction {
	switch d {
	case Up:
		return Down
	case Down:
		return Up
	case Left:
		return Right
	case Right:
		return Left
	case UpperRight:
		return LowerLeft
	case UpperLeft:
		return LowerRight
	case LowerRight:
		return UpperLeft
	case LowerLeft:
		return UpperRight
	}
	return ""
}

func (g *graph) getPath(from gridCoord, to gridCoord, prevSteps []gridCoord) []gridCoord {
	// Figure out what path the arrow should take by traversing the grid recursively.
	var nextPos gridCoord
	log.Debugf("Looking for path from %v to %v", from, to)

	if from == to {
		return []gridCoord{}
	}

	deltaX := to.x - from.x
	deltaY := to.y - from.y

	absX := Abs(deltaX)
	absY := Abs(deltaY)
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
			if g.isFreeInGrid(gridCoord{x: from.x, y: from.y + 1}) && !hasStepBeenTaken(gridCoord{x: from.x, y: from.y + 1}, prevSteps) {
				nextPos = gridCoord{x: from.x, y: from.y + 1}
			} else if g.isFreeInGrid(gridCoord{x: from.x, y: from.y - 1}) && !hasStepBeenTaken(gridCoord{x: from.x, y: from.y - 1}, prevSteps) {
				nextPos = gridCoord{x: from.x, y: from.y - 1}
			} else if g.isFreeInGrid(gridCoord{x: from.x + 1, y: from.y}) && !hasStepBeenTaken(gridCoord{x: from.x + 1, y: from.y}, prevSteps) {
				// Vertical is blocked, let's try horizontal
				nextPos = gridCoord{x: from.x + 1, y: from.y}
			} else {
				// Last resort, try left
				// TODO: Diagonal?
				nextPos = gridCoord{x: from.x - 1, y: from.y}
			}
		} else if graphDirection == "TD" {
			// If we're in TD, we first go horizontal then vertical.
			if g.isFreeInGrid(gridCoord{x: from.x + 1, y: from.y}) && !hasStepBeenTaken(gridCoord{x: from.x + 1, y: from.y}, prevSteps) {
				nextPos = gridCoord{x: from.x + 1, y: from.y}
			} else if g.isFreeInGrid(gridCoord{x: from.x - 1, y: from.y}) && !hasStepBeenTaken(gridCoord{x: from.x - 1, y: from.y}, prevSteps) {
				nextPos = gridCoord{x: from.x - 1, y: from.y}
			} else if g.isFreeInGrid(gridCoord{x: from.x, y: from.y + 1}) && !hasStepBeenTaken(gridCoord{x: from.x, y: from.y + 1}, prevSteps) {
				// Horizontal is blocked, let's try vertical
				nextPos = gridCoord{x: from.x, y: from.y + 1}
			} else {
				// Last resort, try going above
				// TODO: Diagonal?
				nextPos = gridCoord{x: from.x, y: from.y - 1}
			}
		}
	}

	return append(g.getPath(nextPos, to, append(prevSteps, nextPos)), nextPos)
}

func (g *graph) isFreeInGrid(c gridCoord) bool {
	return g.grid[c] == nil
}

func hasStepBeenTaken(step gridCoord, steps []gridCoord) bool {
	for _, s := range steps {
		if s == step {
			return true
		}
	}
	return false
}

func (g *graph) drawArrow(from gridCoord, to gridCoord, label string) {
	dir := determineDirection(genericCoord(from), genericCoord(to))
	log.Debugf("Drawing arrow from %v to %v with direction %s", from, to, dir)

	path := g.getPath(from, to, []gridCoord{})
	d := g.drawing
	previousCoord := from
	var previousDrawingCoord drawingCoord
	for _, nextCoord := range path {
		dir = determineDirection(genericCoord(previousCoord), genericCoord(nextCoord))
		oppositeDir := dir.getOpposite()
		previousDrawingCoord = g.gridToDrawingCoord(previousCoord, &dir)
		nextDrawingCoord := g.gridToDrawingCoord(nextCoord, &oppositeDir)
		log.Debugf("Instructing to draw line from %v to %v (grid %v)", previousDrawingCoord, nextDrawingCoord, nextCoord)
		d.drawLine(previousDrawingCoord, nextDrawingCoord, 0, 0)
		previousCoord = nextCoord
	}
}
