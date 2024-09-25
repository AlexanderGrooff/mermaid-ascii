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

type arrow struct {
	edge *edge
	startDirection direction // Side of the box where the arrow starts from
	endDirection direction // Side of the box where the arrow should end
	coords []coord
}

func (g *graph) getArrowPath(arrow *arrow) []coord {
	// Figure out what path the arrow should take. This takes into account the start node, end node, starting direction and ending direction.
	// Next to this, it should not collide with any nodes in the mapping.

	// If we're in LR, we first go horizontal then vertical.
	// If we're in TD, we first go vertical then horizontal.
}


func (d *drawing) drawArrow(from coord, to coord, label string) {
	dir := determineDirection(from, to)
	log.Debugf("Drawing arrow from %v to %v with direction %s", from, to, dir)

	// Draw arrow body. This lines up between the coords, so the actual from/to
	// coords are offset by 1.
	diffY := Abs(to.y - from.y)
	diffX := Abs(to.x - from.x)
	yLongerThanX := diffY - diffX
	var textCoord coord
	log.Debug("yLongerThanXY: ", diffY, " yLongerThanXX: ", diffX, " yLongerThanX: ", yLongerThanX)
	switch dir {
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
	// If yLongerThanX is positive, we need to draw a vertical line first
	case LowerRight:
		if yLongerThanX == 0 {
			(*d).drawLine(from, to, 1, -1)
			textCoord = coord{from.x + (diffX / 2) - (len(label) / 2), from.y + (diffY / 2)}
		} else {
			var corner coord
			if yLongerThanX > 0 {
				corner = coord{from.x, from.y + yLongerThanX + 2}
				textCoord = coord{Max(from.x-(len(label)/2), 0), from.y + yLongerThanX}
			} else {
				corner = coord{from.x + (diffX + yLongerThanX), to.y}
				if Abs(yLongerThanX) > len(label)+4 {
					textCoord = coord{from.x + (diffY) + Abs(yLongerThanX/2) - (len(label) / 2) - 2, to.y}
				} else {
					textCoord = coord{Max(from.x+diffY-(len(label)/2)-2, 0), Max(corner.y-2, from.y+1)}
				}
			}
			(*d).drawLine(from, corner, 1, -1)
			(*d).drawLine(corner, to, -1, -1)
		}
	case LowerLeft:
		if yLongerThanX == 0 {
			(*d).drawLine(from, to, 1, -1)
			textCoord = coord{from.x - (diffX / 2) - (len(label) / 2), from.y + (diffY / 2)}
		} else {
			var corner coord
			if yLongerThanX > 0 {
				corner = coord{from.x, from.y + yLongerThanX + 2}
				textCoord = coord{from.x - (len(label) / 2), from.y + yLongerThanX}
			} else {
				corner = coord{to.x - yLongerThanX, to.y}
				if Abs(yLongerThanX) > len(label)+4 {
					textCoord = coord{from.x - (diffY) - Abs(yLongerThanX/2) - (len(label) / 4), to.y}
				} else {
					textCoord = coord{from.x - diffY + (len(label) / 2) - 4, Max(corner.y-2, from.y+1)}
				}
			}
			(*d).drawLine(from, corner, 1, -1)
			(*d).drawLine(corner, to, -1, -1)
		}
	case UpperRight:
		if yLongerThanX == 0 {
			corner1 := coord{from.x + 1, from.y}
			corner2 := coord{to.x, to.y + 1}
			(*d).drawLine(from, corner1, 0, 0)
			(*d).drawLine(corner1, corner2, 0, 0)
			(*d).drawLine(corner2, to, 0, -2)
			textCoord = coord{Max(from.x+(diffX/2)-(len(label)/2), 0), from.y - (diffY / 2)}
		} else {
			var corner coord
			if yLongerThanX > 0 {
				corner = coord{to.x, to.y + yLongerThanX - 1}
				textCoord = coord{corner.x - len(label) - 1, corner.y + 1 + len(label)/2}
			} else {
				corner = coord{from.x - yLongerThanX + 1, from.y}
				textCoord = coord{Min(corner.x-len(label)-1, from.x), corner.y}
			}
			(*d).drawLine(from, corner, 1, 0)
			(*d).drawLine(corner, to, 0, -1)
		}
	case UpperLeft:
		if yLongerThanX == 0 {
			(*d).drawLine(from, to, 1, -1)
			textCoord = coord{from.x - (diffX / 2) - (len(label) / 2), from.y - (diffY / 2)}
		} else {
			var corner coord
			if yLongerThanX > 0 {
				corner = coord{from.x, from.y - yLongerThanX}
				textCoord = coord{from.x - (len(label) / 2), from.y - yLongerThanX}
			} else {
				corner = coord{to.x - yLongerThanX + 1, to.y}
				if Abs(yLongerThanX) > len(label)+4 {
					textCoord = coord{from.x - (diffY) - Abs(yLongerThanX/2) - (len(label) / 4), to.y}
				} else {
					textCoord = coord{from.x - diffY + (len(label) / 2) - 4, Max(corner.y+2, from.y-1)}
				}
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
		if dir == UpperRight {
			(*d)[to.x][to.y+1] = "^"
		} else {
			(*d)[to.x-1][to.y] = ">"
		}
	} else {
		// Left
		(*d)[to.x+1][to.y] = "<"
	}
}
