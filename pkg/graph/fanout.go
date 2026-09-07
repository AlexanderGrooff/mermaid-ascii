package graph

import (
	"slices"

	"github.com/mattn/go-runewidth"
)

func (g *graph) layoutFanouts() {
	if g.graphDirection != "TD" || len(g.subgraphs) != 0 {
		return
	}
	for _, n := range g.nodes {
		branches := g.fanoutBranches(n)
		if len(branches) < 2 {
			continue
		}
		row := n.gridCoord.y + 3
		g.rowHeight[row] = Max(g.rowHeight[row], 6)
		from := n.gridCoord.Direction(Down)
		for _, e := range branches {
			to := e.to.gridCoord.Direction(Up)
			branch := gridCoord{x: to.x, y: row}
			if from.x == to.x {
				e.path = []gridCoord{from, to}
			} else {
				e.path = []gridCoord{from, {x: from.x, y: row}, branch, to}
			}
			e.startDir, e.endDir = Down, Up
			e.labelLine = []gridCoord{branch, to}
			e.fanout = true
		}
		slices.SortFunc(branches, func(a, b *edge) int { return a.to.gridCoord.x - b.to.gridCoord.x })
		for i, e := range branches {
			start, _, _ := g.fanoutLabelBounds(e)
			if i == 0 {
				g.offsetX += Max(0, -start)
				continue
			}
			_, previousEnd, _ := g.fanoutLabelBounds(branches[i-1])
			g.columnWidth[e.to.gridCoord.x-1] += Max(0, previousEnd+3-start)
		}
	}
}

func (g *graph) fanoutBranches(n *node) []*edge {
	var branches []*edge
	labeled := false
	seen := make(map[*node]bool)
	for _, e := range g.edges {
		if e.from != n {
			continue
		}
		if e.isBidirectional || e.stroke != strokeSolid || e.head != headArrow || e.to.gridCoord.y != n.gridCoord.y+4 || seen[e.to] {
			return nil
		}
		seen[e.to] = true
		labeled = labeled || e.text != ""
		branches = append(branches, e)
	}
	if !labeled {
		return nil
	}
	row := n.gridCoord.y + 3
	for _, e := range g.edges {
		if e.from == n {
			continue
		}
		fromY, toY := e.from.gridCoord.y, e.to.gridCoord.y
		if rangesOverlap(fromY, toY, row, row) || (fromY == toY && rangesOverlap(fromY-1, fromY+3, row, row)) {
			return nil
		}
	}
	return branches
}

func (g *graph) edgeDrawingCoord(e *edge, c gridCoord) drawingCoord {
	position := g.gridToDrawingCoord(c, nil)
	if e.fanout && c.y == e.from.gridCoord.y+3 {
		position.y = g.gridToDrawingCoord(e.from.gridCoord.Direction(Down), nil).y + 2
	}
	return position
}

func (g *graph) fanoutLabelBounds(e *edge) (int, int, int) {
	from := g.edgeDrawingCoord(e, e.labelLine[0])
	to := g.edgeDrawingCoord(e, e.labelLine[1])
	width := Max(1, runewidth.StringWidth(e.text))
	start := to.x - width/2
	return start, start + width - 1, from.y + (to.y-from.y)/2
}
