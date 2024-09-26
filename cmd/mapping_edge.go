package cmd

type edge struct {
	from *node
	to   *node
	text string
}

func getArrowStartEndOffset(from *node, to *node) (drawingCoord, drawingCoord) {
	// Find which sides the arrow should start/end.
	// This is the middle of one of the sides, depending on the direction of the arrow.
	// Note that the coord returned is relative to the box.
	fromBoxWidth, fromBoxHeight := getDrawingSize(from.drawing)
	toBoxWidth, toBoxHeight := getDrawingSize(to.drawing)
	if from.drawingCoord.x == to.drawingCoord.x {
		// Vertical arrow
		if from.drawingCoord.y < to.drawingCoord.y {
			// Down
			return drawingCoord{fromBoxWidth / 2, fromBoxHeight}, drawingCoord{toBoxWidth / 2, 0}
		} else {
			// Up
			return drawingCoord{fromBoxWidth / 2, 0}, drawingCoord{toBoxWidth / 2, toBoxHeight}
		}
	} else if from.drawingCoord.y == to.drawingCoord.y {
		// Horizontal arrow
		if from.drawingCoord.x < to.drawingCoord.x {
			// Right
			return drawingCoord{fromBoxWidth, fromBoxHeight / 2}, drawingCoord{0, toBoxHeight / 2}
		} else {
			// Left
			return drawingCoord{0, fromBoxHeight / 2}, drawingCoord{toBoxWidth, toBoxHeight / 2}
		}
	} else {
		// Diagonal arrow
		if from.drawingCoord.x < to.drawingCoord.x {
			// Right
			if from.drawingCoord.y < to.drawingCoord.y {
				// Down
				return drawingCoord{fromBoxWidth / 2, fromBoxHeight}, drawingCoord{0, toBoxHeight / 2}
			} else {
				// Up
				return drawingCoord{fromBoxWidth + 1, fromBoxHeight / 2}, drawingCoord{toBoxWidth / 2, toBoxHeight}
			}
		} else {
			// Left
			if from.drawingCoord.y < to.drawingCoord.y {
				// Down
				return drawingCoord{fromBoxWidth / 2, fromBoxHeight}, drawingCoord{toBoxWidth, toBoxHeight / 2}
			} else {
				// Up
				return drawingCoord{fromBoxWidth / 2, 0}, drawingCoord{toBoxWidth, toBoxHeight / 2}
			}
		}
	}
}
