package cmd

type direction genericCoord

var (
	Up         = direction{1, 0}
	Down       = direction{1, 2}
	Left       = direction{0, 1}
	Right      = direction{2, 1}
	UpperRight = direction{2, 0}
	UpperLeft  = direction{0, 0}
	LowerRight = direction{2, 2}
	LowerLeft  = direction{0, 2}
	Middle     = direction{1, 1}
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
	case Middle:
		return Middle
	}
	panic("Unknown direction")
}

func (c gridCoord) Direction(dir direction) gridCoord {
	return gridCoord{x: c.x + dir.x, y: c.y + dir.y}
}
func (c drawingCoord) Direction(dir direction) drawingCoord {
	return drawingCoord{x: c.x + dir.x, y: c.y + dir.y}
}

func selfReferenceDirection(e *edge) (direction, direction, direction, direction) {
	if graphDirection == "LR" {
		return Right, Down, Down, Right
	}
	return Down, Right, Right, Down
}

func determineStartAndEndDir(e *edge) (direction, direction, direction, direction) {
	if e.from == e.to {
		return selfReferenceDirection(e)
	}
	d := determineDirection(genericCoord(*e.from.gridCoord), genericCoord(*e.to.gridCoord))
	var preferredDir, preferredOppositeDir, alternativeDir, alternativeOppositeDir direction

	// Check if this is a backwards flowing edge
	isBackwards := false
	if graphDirection == "LR" {
		// In LR mode, backwards flow is when edge goes from right to left (Left direction)
		isBackwards = (d == Left || d == UpperLeft || d == LowerLeft)
	} else { // TD mode
		// In TD mode, backwards flow is when edge goes from bottom to top (Up direction)
		isBackwards = (d == Up || d == UpperLeft || d == UpperRight)
	}

	// LR: prefer vertical over horizontal
	// TD: prefer horizontal over vertical
	// TODO: This causes some squirmy lines if the corner spot is already occupied.
	// For backwards edges, use special start positions: Down in LR mode, Right in TD mode
	switch d {
	case LowerRight:
		if graphDirection == "LR" {
			preferredDir = Down
			preferredOppositeDir = Left
			alternativeDir = Right
			alternativeOppositeDir = Up
		} else {
			preferredDir = Right
			preferredOppositeDir = Up
			alternativeDir = Down
			alternativeOppositeDir = Left
		}
	case UpperRight:
		if graphDirection == "LR" {
			preferredDir = Up
			preferredOppositeDir = Left
			alternativeDir = Right
			alternativeOppositeDir = Down
		} else {
			preferredDir = Right
			preferredOppositeDir = Down
			alternativeDir = Up
			alternativeOppositeDir = Left
		}
	case LowerLeft:
		if graphDirection == "LR" {
			// Backwards flow in LR mode - start from Down, arrive at Down
			preferredDir = Down
			preferredOppositeDir = Down // Edge goes to bottom of destination
			alternativeDir = Left
			alternativeOppositeDir = Up
		} else {
			preferredDir = Left
			preferredOppositeDir = Up
			alternativeDir = Down
			alternativeOppositeDir = Right
		}
	case UpperLeft:
		if graphDirection == "LR" {
			// Backwards flow in LR mode - start from Down, arrive at Down
			preferredDir = Down
			preferredOppositeDir = Down // Edge goes to bottom of destination
			alternativeDir = Left
			alternativeOppositeDir = Down
		} else {
			// Backwards flow in TD mode - start from Right, arrive at Right
			preferredDir = Right
			preferredOppositeDir = Right // Edge goes to right of destination
			alternativeDir = Up
			alternativeOppositeDir = Right
		}
	default:
		// Handle direct backwards flow cases
		if isBackwards {
			if graphDirection == "LR" && d == Left {
				// Direct left flow in LR mode - start from Down, arrive at Down
				preferredDir = Down
				preferredOppositeDir = Down // Edge goes to bottom of destination
				alternativeDir = Left
				alternativeOppositeDir = Right
			} else if graphDirection == "TD" && d == Up {
				// Direct up flow in TD mode - start from Right, arrive at Right
				preferredDir = Right
				preferredOppositeDir = Right // Edge goes to right of destination
				alternativeDir = Up
				alternativeOppositeDir = Down
			} else {
				preferredDir = d
				preferredOppositeDir = preferredDir.getOpposite()
				alternativeDir = d
				alternativeOppositeDir = preferredOppositeDir
			}
		} else {
			preferredDir = d
			preferredOppositeDir = preferredDir.getOpposite()
			// TODO: just return null and don't calculate alternative path
			alternativeDir = d
			alternativeOppositeDir = preferredOppositeDir
		}
	}
	return preferredDir, preferredOppositeDir, alternativeDir, alternativeOppositeDir
}
