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

func (d *drawing) drawArrow(from coord, to coord, label string) {
	log.Debug("Drawing arrow from ", from, " to ", to)

	// Draw arrow body. This lines up between the coords, so the actual from/to
	// coords are offset by 1.
	diffY := Abs(to.y - from.y)
	diffX := Abs(to.x - from.x)
	diff := diffY - diffX
	var textCoord coord
	log.Debug("diffY: ", diffY, " diffX: ", diffX, " diff: ", diff)
	switch determineDirection(from, to) {
	case Up:
		(*d).drawLine(from, to, 1, -1)
		textCoord = coord{to.x - (len(label) / 2), from.y - (diffY / 2)}
	case Down:
		(*d).drawLine(from, to, 1, -1)
		textCoord = coord{to.x - (len(label) / 2), from.y + (diffY / 2)}
	case Left:
		(*d).drawLine(from, to, 1, -1)
		textCoord = coord{to.x + (diffX / 2) - (len(label) / 2) + 1, to.y}
	case Right:
		(*d).drawLine(from, to, 1, -1)
		textCoord = coord{from.x + (diffX / 2) - (len(label) / 2), to.y}
	// Draw diagonal if the arrow is going from one corner to another.
	// If there can't be a straight diagonal, first draw a vertical or
	// horizontal line to the corner, then draw the diagonal.
	// Draw straight line until we can make a straight diagonal
	// If diff is positive, we need to draw a vertical line first
	case LowerRight:
		if diff == 0 {
			(*d).drawLine(from, to, 1, -1)
		} else {
			var corner coord
			if diff > 0 {
				corner = coord{from.x, from.y + diff + 2}
			} else {
				corner = coord{from.x + (diffX + diff), to.y}
			}
			(*d).drawLine(from, corner, 1, -1)
			(*d).drawLine(corner, to, -1, -1)
		}
	case LowerLeft:
		if diff == 0 {
			(*d).drawLine(from, to, 1, -1)
		} else {
			var corner coord
			if diff > 0 {
				corner = coord{from.x, from.y + diff + 2}
			} else {
				corner = coord{to.x - diff, to.y}
			}
			(*d).drawLine(from, corner, 1, -1)
			(*d).drawLine(corner, to, -1, -1)
		}
	case UpperRight:
		if diff == 0 {
			(*d).drawLine(from, to, 1, -1)
		} else {
			var corner coord
			if diff > 0 {
				corner = coord{from.x, from.y - diff}
			} else {
				corner = coord{to.x + diff - 1, to.y}
			}
			(*d).drawLine(from, corner, 1, 0)
			(*d).drawLine(corner, to, 0, -1)
		}
	case UpperLeft:
		if diff == 0 {
			(*d).drawLine(from, to, 1, -1)
		} else {
			var corner coord
			if diff > 0 {
				corner = coord{from.x, from.y - diff}
			} else {
				corner = coord{to.x - diff + 1, to.y}
			}
			(*d).drawLine(from, corner, 1, 0)
			(*d).drawLine(corner, to, 0, -1)
		}
	}
	if label != "" {
		(*d).drawText(textCoord, label)
	}

	// Draw arrow head depending on direction
	if from.x == to.x {
		// Vertical arrow
		if from.y < to.y {
			// Down
			(*d)[to.x][to.y-1] = "v"
		} else {
			// Up
			(*d)[to.x][to.y+1] = "^"
		}
	} else if from.x < to.x {
		// Right
		(*d)[to.x-1][to.y] = ">"
	} else {
		// Left
		(*d)[to.x+1][to.y] = "<"
	}
}
