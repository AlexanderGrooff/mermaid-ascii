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

func drawArrow(from coord, to coord) drawing {
	arrowDrawing := mkDrawing(Max(from.x, to.x), Max(from.y, to.y))
	log.Debug("Drawing arrow from ", from, " to ", to)

	// Draw arrow body. This lines up between the coords, so the actual from/to
	// coords are offset by 1.
	switch determineDirection(from, to) {
	case Up:
		for y := from.y - 1; y > to.y; y-- {
			arrowDrawing[from.x][y] = "|"
		}
	case Down:
		for y := from.y + 1; y < to.y; y++ {
			arrowDrawing[from.x][y] = "|"
		}
	case Left:
		for x := from.x - 1; x > to.x; x-- {
			arrowDrawing[x][from.y] = "-"
		}
	case Right:
		arrowDrawing.drawLine(coord{from.x + 1, from.y}, coord{to.x - 1, to.y})
	// Draw diagonal if the arrow is going from one corner to another.
	// If there can't be a straight diagonal, first draw a vertical or
	// horizontal line to the corner, then draw the diagonal.
	case LowerRight:
		diff := Abs(to.y-from.y) - Abs(to.x-from.x)
		if diff == 0 {
			arrowDrawing.drawLine(coord{from.x, from.y + 1}, coord{to.x - 1, to.y})
		} else {
			// Draw straight line until we can make a straight diagonal
			// If diff is positive, we need to draw a vertical line first
			if diff > 0 {
				// Draw vertical line until diff is 0
				arrowDrawing.drawLine(coord{from.x, from.y + 1}, coord{from.x, from.y + diff})
				// Draw diagonal
				arrowDrawing.drawLine(coord{from.x, from.y + diff + 1}, coord{to.x - 1, to.y})
			} else {
				// Draw diagonal until we can draw a horizontal line
				arrowDrawing.drawLine(coord{from.x, from.y + 1}, coord{to.x + diff, to.y})
				// Draw horizontal line
				arrowDrawing.drawLine(coord{to.x + diff, to.y}, coord{to.x - 1, to.y})
			}
		}
	}

	// Draw arrow head depending on direction
	if from.x == to.x {
		// Vertical arrow
		if from.y < to.y {
			// Down
			arrowDrawing[to.x][to.y-1] = "v"
		} else {
			// Up
			arrowDrawing[to.x][to.y+1] = "^"
		}
	} else if from.x < to.x {
		// Right
		arrowDrawing[to.x-1][to.y] = ">"
	} else {
		// Left
		arrowDrawing[to.x+1][to.y] = "<"
	}

	return arrowDrawing
}
