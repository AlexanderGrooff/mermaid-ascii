package er

import (
	"math"
	"strings"

	"github.com/mattn/go-runewidth"
)

// canvas is a growable 2D grid of runes that boxes are stamped onto and
// connectors are drawn across.
type canvas struct {
	rows [][]rune
}

func (c *canvas) ensure(x, y int) {
	for len(c.rows) <= y {
		c.rows = append(c.rows, nil)
	}
	for len(c.rows[y]) <= x {
		c.rows[y] = append(c.rows[y], ' ')
	}
}

func (c *canvas) set(x, y int, r rune) {
	if x < 0 || y < 0 {
		return
	}
	c.ensure(x, y)
	c.rows[y][x] = r
}

func (c *canvas) at(x, y int) rune {
	if y < 0 || y >= len(c.rows) || x < 0 || x >= len(c.rows[y]) {
		return ' '
	}
	return c.rows[y][x]
}

// stamp places a block of pre-rendered lines with its top-left at (x0,y0).
func (c *canvas) stamp(x0, y0 int, block []string) {
	for dy, line := range block {
		x := x0
		for _, r := range line {
			c.set(x, y0+dy, r)
			x++
		}
	}
}

func (c *canvas) String() string {
	var b strings.Builder
	for _, row := range c.rows {
		b.WriteString(strings.TrimRight(string(row), " "))
		b.WriteByte('\n')
	}
	return b.String()
}

// placedEntity is an entity's rendered box positioned on the canvas.
type placedEntity struct {
	entity *Entity
	lines  []string
	x, y   int // top-left
	w, h   int // box dimensions
}

// centre-ish anchor points on each side of the box, used to attach connectors.
func (p *placedEntity) leftAnchor() (int, int)   { return p.x, p.y + p.h/2 }
func (p *placedEntity) rightAnchor() (int, int)  { return p.x + p.w - 1, p.y + p.h/2 }
func (p *placedEntity) topAnchor() (int, int)    { return p.x + p.w/2, p.y }
func (p *placedEntity) bottomAnchor() (int, int) { return p.x + p.w/2, p.y + p.h - 1 }

const (
	hGap = 6 // horizontal space between grid columns (room for connectors)
	vGap = 3 // vertical space between grid rows
)

// placeEntities renders every entity and arranges the boxes in a near-square
// grid, returning their positions keyed by entity name.
func placeEntities(d *ErDiagram, g glyphs) (map[string]*placedEntity, []*placedEntity) {
	n := len(d.Entities)
	cols := int(math.Ceil(math.Sqrt(float64(n))))
	if cols < 1 {
		cols = 1
	}
	rows := (n + cols - 1) / cols

	// Render boxes and record sizes.
	placed := make([]*placedEntity, n)
	for i, e := range d.Entities {
		lines := renderEntity(e, g)
		placed[i] = &placedEntity{entity: e, lines: lines, w: blockWidth(lines), h: len(lines)}
	}

	// Per-column max width and per-row max height.
	colW := make([]int, cols)
	rowH := make([]int, rows)
	for i, p := range placed {
		r, c := i/cols, i%cols
		if p.w > colW[c] {
			colW[c] = p.w
		}
		if p.h > rowH[r] {
			rowH[r] = p.h
		}
	}

	// Column x-offsets and row y-offsets.
	colX := make([]int, cols)
	x := 0
	for c := 0; c < cols; c++ {
		colX[c] = x
		x += colW[c] + hGap
	}
	rowY := make([]int, rows)
	y := 0
	for r := 0; r < rows; r++ {
		rowY[r] = y
		y += rowH[r] + vGap
	}

	byName := map[string]*placedEntity{}
	for i, p := range placed {
		r, c := i/cols, i%cols
		p.x, p.y = colX[c], rowY[r]
		byName[p.entity.Name] = p
	}
	return byName, placed
}

func blockWidth(lines []string) int {
	w := 0
	for _, l := range lines {
		if lw := runewidth.StringWidth(l); lw > w {
			w = lw
		}
	}
	return w
}
