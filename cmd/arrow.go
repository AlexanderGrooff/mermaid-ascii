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

func determineDirection(from coord, to coord) direction {
	if from.x == to.x {
		if from.y < to.y {
			return Down
		} else {
			return Up
		}
	} else if from.y == to.y {
		if from.x < to.x {
			return Right
		} else {
			return Left
		}
	} else if from.x < to.x {
		if from.y < to.y {
			return LowerRight
		} else {
			return UpperRight
		}
	} else {
		if from.y < to.y {
			return LowerLeft
		} else {
			return UpperLeft
		}
	}
}

func drawArrow(from coord, to coord) drawing {
	// Stop arrow one character before the end coord to stop just before the target
	arrowDrawing := mkDrawing(Max(from.x, to.x), Max(from.y, to.y))
	log.Debug("Drawing arrow from ", from, " to ", to)

	// Draw arrow body
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
		for x := from.x + 1; x < to.x; x++ {
			arrowDrawing[x][from.y] = "-"
		}
	// Draw diagonal if the arrow is going from one corner to another.
	// If there can't be a straight diagonal, first draw a vertical or
	// horizontal line to the corner, then draw the diagonal.
	case LowerRight:
		diff := Abs(to.y-from.y) - Abs(to.x-from.x)
		log.Debug("Drawing diagonal with diff ", diff)
		if diff == 0 {
			for x, y := from.x, from.y+1; x < to.x && y < to.y; x, y = x+1, y+1 {
				arrowDrawing[x][y] = "\\"
			}
		} else {
			// Draw straight line until we can make a straight diagonal
			// If diff is positive, we need to draw a vertical line first
			if diff > 0 {
				// Draw vertical line until diff is 0
				for y := from.y + 1; y < from.y+1+diff; y++ {
					arrowDrawing[from.x][y] = "|"
				}
				// Draw diagonal
				for x, y := from.x, from.y+1+diff; x < to.x && y < to.y; x, y = x+1, y+1 {
					arrowDrawing[x][y] = "\\"
				}
			} else {
				// Draw diagonal until we can draw a horizontal line
				for x, y := from.x, from.y+1; x < to.x && y < to.y; x, y = x+1, y+1 {
					arrowDrawing[x][y] = "\\"
				}
				// Draw horizontal line
				for x := to.x + diff - 1; x < to.x; x++ {
					arrowDrawing[x][to.y] = "-"
				}
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
