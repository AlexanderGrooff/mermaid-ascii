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

	// Widen the horizontal gap so relationship labels + crow's-foot tokens fit
	// between columns without spilling into a box.
	gap := hGap
	for _, r := range d.Relationships {
		if need := len([]rune(r.Label)) + 8; need > gap {
			gap = need
		}
	}

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
		x += colW[c] + gap
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

// crow's-foot cardinality markers, as read toward the box. The tokens match
// mermaid's own notation (|| one, |o/o| zero-or-one, }o/o{ zero-or-more,
// }|/|{ one-or-more).
func leftToken(c Cardinality) string {
	switch c {
	case OnlyOne:
		return "||"
	case ZeroOrOne:
		return "|o"
	case ZeroOrMore:
		return "}o"
	default: // OneOrMore
		return "}|"
	}
}

func rightToken(c Cardinality) string {
	switch c {
	case OnlyOne:
		return "||"
	case ZeroOrOne:
		return "o|"
	case ZeroOrMore:
		return "o{"
	default: // OneOrMore
		return "|{"
	}
}

// drawConnectors routes each relationship between its two boxes with an
// orthogonal line, crow's-foot markers at each end, and the label at the mid.
func drawConnectors(c *canvas, byName map[string]*placedEntity, d *ErDiagram, g glyphs) {
	for _, r := range d.Relationships {
		a, b := byName[r.Left], byName[r.Right]
		if a == nil || b == nil {
			continue
		}
		lineChar := g.h
		vChar := g.v
		if !r.Identifying {
			lineChar, vChar = ':', ':' // dashed (non-identifying) approximation
		}

		// Order left→right by x so tokens attach to the correct sides.
		la, lb := a, b
		lt, rt := leftToken(r.LeftCard), rightToken(r.RightCard)
		if b.x < a.x {
			la, lb = b, a
			lt, rt = rightTokenFlip(r.RightCard), leftTokenFlip(r.LeftCard)
		}

		ax, ay := la.rightAnchor()
		bx, by := lb.leftAnchor()
		if lb.x <= la.x+la.w { // not cleanly to the right → use vertical routing
			drawVertical(c, la, lb, r, g)
			continue
		}

		// Horizontal orthogonal route: A.right → midX → B.left, with a vertical
		// jog at midX when the two anchors sit on different rows.
		midX := (ax + bx) / 2
		startX := ax + 1 + len([]rune(lt)) // leave room for the left token
		for x := startX; x < midX; x++ {
			c.set(x, ay, lineChar)
		}
		for y := minI(ay, by); y <= maxI(ay, by); y++ {
			c.set(midX, y, vChar)
		}
		for x := midX; x < bx-len([]rune(rt)); x++ {
			c.set(x, by, lineChar)
		}
		// Crow's-foot tokens hugging each box.
		putStr(c, ax+1, ay, lt)
		putStr(c, bx-len([]rune(rt)), by, rt)
		// Label centred on the line, clipped so it never enters box B.
		if r.Label != "" {
			lx := midX - len([]rune(r.Label))/2
			if lx < startX {
				lx = startX
			}
			putStrClip(c, lx, by, r.Label, bx-len([]rune(rt))-1)
		}
	}
}

// putStrClip writes s at (x,y) but stops before column maxX (exclusive).
func putStrClip(c *canvas, x, y int, s string, maxX int) {
	for _, r := range s {
		if x >= maxX {
			return
		}
		c.set(x, y, r)
		x++
	}
}

// drawVertical routes a relationship whose boxes are stacked (or overlapping in
// x) with a vertical orthogonal line between top/bottom anchors.
func drawVertical(c *canvas, a, b *placedEntity, r *Relationship, g glyphs) {
	top, bot := a, b
	tt, bt := leftToken(r.LeftCard), rightToken(r.RightCard)
	if b.y < a.y {
		top, bot = b, a
		tt, bt = rightToken(r.RightCard), leftToken(r.LeftCard)
	}
	tx, ty := top.bottomAnchor()
	bx, by := bot.topAnchor()
	vChar := g.v
	if !r.Identifying {
		vChar = ':'
	}
	midY := (ty + by) / 2
	for y := ty + 1 + 1; y < midY; y++ {
		c.set(tx, y, vChar)
	}
	for x := minI(tx, bx); x <= maxI(tx, bx); x++ {
		c.set(x, midY, g.h)
	}
	for y := midY; y < by-1; y++ {
		c.set(bx, y, vChar)
	}
	putStr(c, tx, ty+1, tt)
	putStr(c, bx, by-1, bt)
	if r.Label != "" {
		putStr(c, maxI(tx, bx)+2, midY, r.Label)
	}
}

// leftTokenFlip/rightTokenFlip give the box-facing token when the visual order
// is reversed (right entity drawn on the left).
func leftTokenFlip(c Cardinality) string  { return rightToken(c) }
func rightTokenFlip(c Cardinality) string { return leftToken(c) }

func putStr(c *canvas, x, y int, s string) {
	for _, r := range s {
		c.set(x, y, r)
		x++
	}
}

func minI(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func maxI(a, b int) int {
	if a > b {
		return a
	}
	return b
}
