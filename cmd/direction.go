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
