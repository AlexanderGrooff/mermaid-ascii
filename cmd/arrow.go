package cmd

import (
	"container/heap"
	"fmt"
	"slices"

	log "github.com/sirupsen/logrus"
)

type priorityQueueItem struct {
	coord    gridCoord
	priority int
	index    int
}

type priorityQueue []*priorityQueueItem

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].priority < pq[j].priority
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*priorityQueueItem)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

func heuristic(a, b gridCoord) int {
	absX := Abs(a.x - b.x)
	absY := Abs(a.y - b.y)
	if absX == 0 || absY == 0 {
		return absX + absY
	} else {
		// Punish for taking extra corner, we prefer straight (less complex) lines
		return absX + absY + 1
	}
}

func (g *graph) getPath(from gridCoord, to gridCoord) ([]gridCoord, error) {
	pq := &priorityQueue{}
	heap.Init(pq)
	heap.Push(pq, &priorityQueueItem{coord: from, priority: 0})

	costSoFar := map[gridCoord]int{from: 0}
	cameFrom := map[gridCoord]*gridCoord{from: nil}

	directions := []gridCoord{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}

	for pq.Len() > 0 {
		current := heap.Pop(pq).(*priorityQueueItem).coord

		if current.Equals(to) {
			path := []gridCoord{}
			for c := &current; c != nil; c = cameFrom[*c] {
				path = append([]gridCoord{*c}, path...)
			}
			log.Debugf("Found it! path: %v", path)
			return path, nil
		}

		for _, dir := range directions {
			next := gridCoord{x: current.x + dir.x, y: current.y + dir.y}
			if !g.isFreeInGrid(next) && !next.Equals(to) {
				continue
			}

			newCost := costSoFar[current] + 1
			if cost, ok := costSoFar[next]; !ok || newCost < cost {
				costSoFar[next] = newCost
				priority := newCost + heuristic(next, to)
				heap.Push(pq, &priorityQueueItem{coord: next, priority: priority})
				cameFrom[next] = &current
			}
		}
	}

	return nil, fmt.Errorf("no path found")
}

func (g *graph) isFreeInGrid(c gridCoord) bool {
	// We'll fix it later if we overshoot the grid size
	if c.x < 0 || c.y < 0 {
		return false
	}
	return g.grid[c] == nil
}

func (g *graph) drawArrow(from gridCoord, to gridCoord, e *edge) (*drawing, *drawing, *drawing, *drawing, *drawing) {
	if len(e.path) == 0 {
		return nil, nil, nil, nil, nil
	}
	log.Debugf("Drawing arrow from %v to %v with path %v", from, to, e.path)
	dLabel := g.drawArrowLabel(e)
	dPath, linesDrawn, lineDirs := g.drawPath(e.path)
	dBoxStart := g.drawBoxStart(e.path, linesDrawn[0])
	dArrowHead := g.drawArrowHead(linesDrawn[len(linesDrawn)-1], lineDirs[len(lineDirs)-1])
	dCorners := g.drawCorners(e.path)
	return dPath, dBoxStart, dArrowHead, dCorners, dLabel
}

func mergePath(path []gridCoord) []gridCoord {
	// If two steps are in the same direction, merge them to one step.
	if len(path) <= 2 {
		return path
	}
	indexToRemove := []int{}
	step0 := path[0]
	step1 := path[1]
	for idx, step2 := range path[2:] {
		prevDir := determineDirection(genericCoord(step0), genericCoord(step1))
		dir := determineDirection(genericCoord(step1), genericCoord(step2))
		if prevDir == dir {
			indexToRemove = append(indexToRemove, idx+1) // +1 because we skip the initial step
		}
		step0 = step1
		step1 = step2
	}
	newPath := []gridCoord{}
	for idx, step := range path {
		if !slices.Contains(indexToRemove, idx) {
			newPath = append(newPath, step)
		}
	}
	log.Debugf("Reduced path from %v to %v", path, newPath)
	return newPath
}

func (g *graph) drawPath(path []gridCoord) (*drawing, [][]drawingCoord, []direction) {
	d := copyCanvas(g.drawing)
	previousCoord := path[0]
	linesDrawn := make([][]drawingCoord, 0)
	lineDirs := make([]direction, 0)
	var previousDrawingCoord drawingCoord
	for _, nextCoord := range path[1:] {
		previousDrawingCoord = g.gridToDrawingCoord(previousCoord, nil)
		nextDrawingCoord := g.gridToDrawingCoord(nextCoord, nil)
		if previousDrawingCoord.Equals(nextDrawingCoord) {
			log.Debugf("Skipping drawing identical line on %v", nextCoord)
			continue
		}
		dir := determineDirection(genericCoord(previousCoord), genericCoord(nextCoord))
		s := d.drawLine(previousDrawingCoord, nextDrawingCoord, 1, -1)
		if len(s) == 0 {
			// drawLine may return no coords if offsets collapse the line. Use at least one point so arrow and junction logic
			// can still infer a direction.
			s = append(s, previousDrawingCoord)
		}
		linesDrawn = append(linesDrawn, s)
		lineDirs = append(lineDirs, dir)
		previousCoord = nextCoord
	}
	return d, linesDrawn, lineDirs
}

func (g *graph) drawBoxStart(path []gridCoord, firstLine []drawingCoord) *drawing {
	d := *(copyCanvas(g.drawing))
	from := firstLine[0]
	dir := determineDirection(genericCoord(path[0]), genericCoord(path[1]))
	log.Debugf("Drawing box start at %v with direction %v for line %v", from, dir, path)

	if useAscii {
		return &d
	}

	switch dir {
	case Up:
		d[from.x][from.y+1] = "┴"
	case Down:
		d[from.x][from.y-1] = "┬"
	case Left:
		d[from.x+1][from.y] = "┤"
	case Right:
		d[from.x-1][from.y] = "├"
	}
	return &d
}

func (g *graph) drawArrowHead(line []drawingCoord, fallback direction) *drawing {
	d := *(copyCanvas(g.drawing))
	if len(line) == 0 {
		return &d
	}
	from := line[0]
	lastPos := line[len(line)-1]
	dir := determineDirection(genericCoord(from), genericCoord(lastPos))
	if len(line) == 1 || dir == Middle {
		dir = fallback
	}

	var char string
	if !useAscii {
		switch dir {
		case Up:
			char = "▲"
		case Down:
			char = "▼"
		case Left:
			char = "◄"
		case Right:
			char = "►"
		case UpperRight:
			char = "◥"
		case UpperLeft:
			char = "◤"
		case LowerRight:
			char = "◢"
		case LowerLeft:
			char = "◣"
		default:
			switch fallback {
			case Up:
				char = "▲"
			case Down:
				char = "▼"
			case Left:
				char = "◄"
			case Right:
				char = "►"
			case UpperRight:
				char = "◥"
			case UpperLeft:
				char = "◤"
			case LowerRight:
				char = "◢"
			case LowerLeft:
				char = "◣"
			default:
				char = "●"
			}
		}
	} else {
		switch dir {
		case Up:
			char = "^"
		case Down:
			char = "v"
		case Left:
			char = "<"
		case Right:
			char = ">"
		default:
			switch fallback {
			case Up:
				char = "^"
			case Down:
				char = "v"
			case Left:
				char = "<"
			case Right:
				char = ">"
			default:
				char = "*"
			}
		}
	}

	d[lastPos.x][lastPos.y] = char
	return &d
}

func (g *graph) drawCorners(path []gridCoord) *drawing {
	d := copyCanvas(g.drawing)
	for idx, coord := range path {
		// Skip the first and last step
		if idx == 0 || idx == len(path)-1 {
			continue
		}
		drawingCoord := g.gridToDrawingCoord(coord, nil)

		prevDir := determineDirection(genericCoord(path[idx-1]), genericCoord(coord))
		nextDir := determineDirection(genericCoord(coord), genericCoord(path[idx+1]))

		var corner string
		if !useAscii {
			switch {
			case (prevDir == Right && nextDir == Down) || (prevDir == Up && nextDir == Left):
				corner = "┐"
			case (prevDir == Right && nextDir == Up) || (prevDir == Down && nextDir == Left):
				corner = "┘"
			case (prevDir == Left && nextDir == Down) || (prevDir == Up && nextDir == Right):
				corner = "┌"
			case (prevDir == Left && nextDir == Up) || (prevDir == Down && nextDir == Right):
				corner = "└"
			default:
				corner = "+"
			}
		} else {
			corner = "+"
		}

		(*d)[drawingCoord.x][drawingCoord.y] = corner
	}
	return d
}

func (g *graph) drawArrowLabel(e *edge) *drawing {
	d := copyCanvas(g.drawing)
	lenLabel := len(e.text)
	if lenLabel == 0 {
		return d
	}

	log.Debugf("Drawing text '%s' on gridline %v", e.text, e.labelLine)
	d.drawTextOnLine(g.lineToDrawing(e.labelLine), e.text)
	return d
}

func (d *drawing) drawTextOnLine(line []drawingCoord, label string) {
	// Write text in middle of the line
	//  123456789
	// |---------|
	//     123
	log.Debugf("Drawing text '%s' on drawingline %v", label, line)
	var minX, maxX, minY, maxY int
	if line[0].x > line[1].x {
		minX = line[1].x
		maxX = line[0].x
	} else {
		minX = line[0].x
		maxX = line[1].x
	}
	if line[0].y > line[1].y {
		minY = line[1].y
		maxY = line[0].y
	} else {
		minY = line[0].y
		maxY = line[1].y
	}
	middleX := minX + (maxX-minX)/2
	middleY := minY + (maxY-minY)/2
	startLabelCoord := drawingCoord{x: middleX - len(label)/2, y: middleY}
	d.drawText(startLabelCoord, label)
}
