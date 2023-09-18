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

func drawArrow(from coord, to coord) *drawing {
	arrowDrawing := *(mkDrawing(Max(from.x, to.x), Max(from.y, to.y)))
	dir := determineDirection(from, to)
	log.Debugf("Drawing arrow from %v to %v with direction %s", from, to, dir)

	// Draw arrow body. This lines up between the coords, so the actual from/to
	// coords are offset by 1.
	diffY := Abs(to.y - from.y)
	diffX := Abs(to.x - from.x)
	yLongerThanX := diffY - diffX
	log.Debug("diffY: ", diffY, " diffX: ", diffX, " diff: ", yLongerThanX)
	switch dir {
	case Up:
		arrowDrawing.drawLine(from, to, 1, -1)
	case Down:
		arrowDrawing.drawLine(from, to, 1, -1)
	case Left:
		arrowDrawing.drawLine(from, to, 1, -1)
	case Right:
		arrowDrawing.drawLine(from, to, 1, -1)
	// Draw diagonal if the arrow is going from one corner to another.
	// If there can't be a straight diagonal, first draw a vertical or
	// horizontal line to the corner, then draw the diagonal.
	// Draw straight line until we can make a straight diagonal
	// If diff is positive, we need to draw a vertical line first
	case LowerRight:
		if yLongerThanX == 0 {
			arrowDrawing.drawLine(from, to, 1, -1)
		} else {
			var corner coord
			if yLongerThanX > 0 {
				corner = coord{from.x, from.y + yLongerThanX + 2}
			} else {
				corner = coord{from.x + (diffX + yLongerThanX), to.y}
			}
			arrowDrawing.drawLine(from, corner, 1, -1)
			arrowDrawing.drawLine(corner, to, -1, -1)
		}
	case LowerLeft:
		if yLongerThanX == 0 {
			arrowDrawing.drawLine(from, to, 1, -1)
		} else {
			var corner coord
			if yLongerThanX > 0 {
				corner = coord{from.x, from.y + yLongerThanX + 2}
			} else {
				corner = coord{to.x - yLongerThanX, to.y}
			}
			arrowDrawing.drawLine(from, corner, 1, -1)
			arrowDrawing.drawLine(corner, to, -1, -1)
		}
	case UpperRight:
		if yLongerThanX == 0 {
			corner1 := coord{from.x + 1, from.y}
			corner2 := coord{to.x, to.y + 1}
			arrowDrawing.drawLine(from, corner1, 0, 0)
			arrowDrawing.drawLine(corner1, corner2, 0, 0)
			arrowDrawing.drawLine(corner2, to, 0, -1)
		} else {
			var corner coord
			if yLongerThanX > 0 {
				corner = coord{to.x, to.y + yLongerThanX - 1}
			} else {
				corner = coord{from.x - yLongerThanX + 1, from.y}
			}
			arrowDrawing.drawLine(from, corner, 1, 0)
			arrowDrawing.drawLine(corner, to, 0, -1)
		}
	case UpperLeft:
		if yLongerThanX == 0 {
			arrowDrawing.drawLine(from, to, 1, -1)
		} else {
			var corner coord
			if yLongerThanX > 0 {
				corner = coord{from.x, from.y - yLongerThanX}
			} else {
				corner = coord{to.x - yLongerThanX + 1, to.y}
			}
			arrowDrawing.drawLine(from, corner, 1, 0)
			arrowDrawing.drawLine(corner, to, 0, -1)
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
		if dir == UpperRight {
			arrowDrawing[to.x][to.y+1] = "^"
		} else {
			arrowDrawing[to.x-1][to.y] = ">"
		}
	} else {
		// Left
		arrowDrawing[to.x+1][to.y] = "<"
	}

	return &arrowDrawing
}
