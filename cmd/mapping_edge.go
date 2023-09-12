package cmd

import (
)

type edge struct {
	from node
	to   node
	text string
	drawn bool
}

func getArrowStartEndOffset(from node, to node) (coord, coord) {
	// Find which sides the arrow should start/end.
	// This is the middle of one of the sides, depending on the direction of the arrow.
	// Note that the coord returned is relative to the box.
	fromBoxWidth, fromBoxHeight := getDrawingSize(from.drawing)
	toBoxWidth, toBoxHeight := getDrawingSize(to.drawing)
	if from.coord.x == to.coord.x {
		// Vertical arrow
		if from.coord.y < to.coord.y {
			// Down
			return coord{fromBoxWidth / 2, fromBoxHeight}, coord{toBoxWidth / 2, 0}
		} else {
			// Up
			return coord{fromBoxWidth / 2, 0}, coord{toBoxWidth / 2, toBoxHeight}
		}
	} else if from.coord.y == to.coord.y {
		// Horizontal arrow
		if from.coord.x < to.coord.x {
			// Right
			return coord{fromBoxWidth, fromBoxHeight / 2}, coord{0, toBoxHeight / 2}
		} else {
			// Left
			return coord{0, fromBoxHeight / 2}, coord{toBoxWidth, toBoxHeight / 2}
		}
	} else {
		// Diagonal arrow
		if from.coord.x < to.coord.x {
			// Right
			if from.coord.y < to.coord.y {
				// Down
				return coord{fromBoxWidth / 2, fromBoxHeight}, coord{0, toBoxHeight / 2}
			} else {
				// Up
				return coord{fromBoxWidth / 2, 0}, coord{0, toBoxHeight / 2}
			}
		} else {
			// Left
			if from.coord.y < to.coord.y {
				// Down
				return coord{fromBoxWidth / 2, fromBoxHeight}, coord{toBoxWidth, toBoxHeight / 2}
			} else {
				// Up
				return coord{fromBoxWidth / 2, 0}, coord{toBoxWidth, toBoxHeight / 2}
			}
		}
	}
}
