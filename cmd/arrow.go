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
	if c.x < 0 || c.y < 0 || c.x >= len(g.columnWidth) || c.y >= len(g.rowHeight) {
		return false
	}
	return g.grid[c] == nil
}

func (g *graph) drawArrow(from gridCoord, to gridCoord, e *edge) {
	log.Debugf("Drawing arrow from %v to %v with path %v", from, to, e.path)
	dLabel := g.drawArrowLabel(e)
	dPath, linesDrawn := g.drawPath(e.path)
	dHead := g.drawArrowHead(linesDrawn[len(linesDrawn)-1])
	g.drawing = mergeDrawings(g.drawing, dPath, drawingCoord{0, 0})
	g.drawing = mergeDrawings(g.drawing, dHead, drawingCoord{0, 0})
	g.drawing = mergeDrawings(g.drawing, dLabel, drawingCoord{0, 0})
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
			log.Debugf("Removing %v from path", step1)
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
	return newPath
}

func (g *graph) drawPath(path []gridCoord) (*drawing, [][]drawingCoord) {
	d := copyCanvas(g.drawing)
	previousCoord := path[0]
	linesDrawn := make([][]drawingCoord, 0)
	var previousDrawingCoord drawingCoord
	for idx, nextCoord := range path[1:] {
		previousDrawingCoord = g.gridToDrawingCoord(previousCoord, nil)
		nextDrawingCoord := g.gridToDrawingCoord(nextCoord, nil)
		if previousDrawingCoord.Equals(nextDrawingCoord) {
			log.Debugf("Skipping drawing identical line on %v", nextCoord)
			continue
		}
		if idx == 0 {
			// Don't cross the node border
			linesDrawn = append(linesDrawn, d.drawLine(previousDrawingCoord, nextDrawingCoord, 1, -1))
		} else {
			linesDrawn = append(linesDrawn, d.drawLine(previousDrawingCoord, nextDrawingCoord, 0, -1))
		}
		previousCoord = nextCoord
	}
	return d, linesDrawn
}

func (g *graph) drawArrowHead(line []drawingCoord) *drawing {
	d := *(copyCanvas(g.drawing))
	// Determine the direction of the arrow for the last step
	from := line[0]
	lastPos := line[len(line)-1]
	dir := determineDirection(genericCoord(from), genericCoord(lastPos))
	switch dir {
	case Up:
		d[lastPos.x][lastPos.y] = "^"
	case Down:
		d[lastPos.x][lastPos.y] = "v"
	case Left:
		d[lastPos.x][lastPos.y] = "<"
	case Right:
		d[lastPos.x][lastPos.y] = ">"
	case UpperRight:
		d[lastPos.x][lastPos.y] = "┐"
	case UpperLeft:
		d[lastPos.x][lastPos.y] = "┌"
	case LowerRight:
		d[lastPos.x][lastPos.y] = "┘"
	case LowerLeft:
		d[lastPos.x][lastPos.y] = "└"
	default:
		d[lastPos.x][lastPos.y] = "+"
	}
	return &d
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
