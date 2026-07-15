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

// side identifies which face of a box a connector attaches to.
type side int

const (
	sideL side = iota
	sideR
	sideT
	sideB
)

// placedEntity is an entity's rendered box positioned on the canvas.
type placedEntity struct {
	entity   *Entity
	lines    []string
	x, y     int // top-left
	w, h     int // box dimensions
	row, col int // grid cell
}

// layout holds the placed boxes plus the gutter geometry connectors route
// through. Every gap between grid cells (and a margin all around) is a routing
// gutter that is `lanes` cells wide/tall, so each relationship can travel in its
// own lane and never overlap another.
type layout struct {
	byName map[string]*placedEntity
	placed []*placedEntity
	lanes  int   // one lane per relationship (global, so lanes never clash)
	vGutX  []int // left edge x of each vertical gutter (len cols+1)
	hGutY  []int // top edge y of each horizontal gutter (len rows+1)
	cols   int
}

// placeEntities renders every entity and arranges the boxes in a near-square
// grid separated by lane-wide gutters.
func placeEntities(d *ErDiagram, g glyphs) *layout {
	n := len(d.Entities)
	cols := int(math.Ceil(math.Sqrt(float64(n))))
	if cols < 1 {
		cols = 1
	}
	rows := (n + cols - 1) / cols

	lanes := len(d.Relationships)
	if lanes < 1 {
		lanes = 1
	}

	// Vertical gutters must be wide enough both for one lane per relationship and
	// to give the widest label clear room to sit on its horizontal run.
	maxLabel := 0
	for _, r := range d.Relationships {
		if w := runewidth.StringWidth(r.Label); w > maxLabel {
			maxLabel = w
		}
	}
	gutW := maxI(lanes, maxLabel) + 5

	placed := make([]*placedEntity, n)
	for i, e := range d.Entities {
		lines := renderEntity(e, g)
		placed[i] = &placedEntity{
			entity: e, lines: lines,
			w: blockWidth(lines), h: len(lines),
			row: i / cols, col: i % cols,
		}
	}

	// Per-column width and per-row height.
	colW := make([]int, cols)
	rowH := make([]int, rows)
	for _, p := range placed {
		if p.w > colW[p.col] {
			colW[p.col] = p.w
		}
		if p.h > rowH[p.row] {
			rowH[p.row] = p.h
		}
	}

	// X layout: [gutter][col0][gutter][col1]…[gutter]. Each gutter is `lanes`
	// wide so every relationship gets a clear vertical lane.
	vGutX := make([]int, cols+1)
	colX := make([]int, cols)
	x := 0
	for c := 0; c < cols; c++ {
		vGutX[c] = x
		x += gutW
		colX[c] = x
		x += colW[c]
	}
	vGutX[cols] = x

	// Y layout mirrors X: [gutter][row0][gutter]…[gutter].
	hGutY := make([]int, rows+1)
	rowY := make([]int, rows)
	y := 0
	for r := 0; r < rows; r++ {
		hGutY[r] = y
		y += lanes
		rowY[r] = y
		y += rowH[r]
	}
	hGutY[rows] = y

	byName := map[string]*placedEntity{}
	for _, p := range placed {
		p.x, p.y = colX[p.col], rowY[p.row]
		byName[p.entity.Name] = p
	}
	return &layout{byName: byName, placed: placed, lanes: lanes, vGutX: vGutX, hGutY: hGutY, cols: cols}
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

// ---- connector routing -------------------------------------------------------

// dir bits mark which neighbours a connector cell links to; the glyph for a cell
// is chosen from the union of its bits (so crossings become ┼, corners └┐┌┘…).
const (
	dN uint8 = 1 << iota
	dS
	dE
	dW
)

// overlay accumulates connector line bits per cell (kept off the box canvas so
// junction glyphs can be computed once at the end). solid vs dashed is tracked
// separately so an identifying line stays solid where it doesn't cross a dashed
// one.
type overlay struct {
	solid map[[2]int]uint8
	dash  map[[2]int]uint8
	label map[[2]int]rune
	token map[[2]int]rune
}

func newOverlay() *overlay {
	return &overlay{
		solid: map[[2]int]uint8{}, dash: map[[2]int]uint8{},
		label: map[[2]int]rune{}, token: map[[2]int]rune{},
	}
}

func (o *overlay) bits(x, y int) uint8 { return o.solid[[2]int{x, y}] | o.dash[[2]int{x, y}] }

// polyline sets connector bits along an axis-aligned poly-line through pts.
func (o *overlay) polyline(pts [][2]int, solid bool) {
	m := o.dash
	if solid {
		m = o.solid
	}
	for i := 0; i+1 < len(pts); i++ {
		a, b := pts[i], pts[i+1]
		dx, dy := sign(b[0]-a[0]), sign(b[1]-a[1])
		x, y := a[0], a[1]
		for {
			if dx > 0 {
				m[[2]int{x, y}] |= dE
			} else if dx < 0 {
				m[[2]int{x, y}] |= dW
			}
			if dy > 0 {
				m[[2]int{x, y}] |= dS
			} else if dy < 0 {
				m[[2]int{x, y}] |= dN
			}
			if x == b[0] && y == b[1] {
				break
			}
			x += dx
			y += dy
			// mark the reverse link on the cell we just entered
			if dx > 0 {
				m[[2]int{x, y}] |= dW
			} else if dx < 0 {
				m[[2]int{x, y}] |= dE
			}
			if dy > 0 {
				m[[2]int{x, y}] |= dN
			} else if dy < 0 {
				m[[2]int{x, y}] |= dS
			}
		}
	}
}

// attach chooses evenly-spaced connection points along each box side so several
// relationships touching the same side don't stack on one cell. It returns the
// (x,y) for the idx-th of total connectors on that side.
func attach(p *placedEntity, s side, idx, total int) (int, int) {
	slot := func(lo, hi int) int {
		span := hi - lo
		if total <= 1 || span <= 0 {
			return lo + span/2
		}
		return lo + span*(idx+1)/(total+1)
	}
	switch s {
	case sideL:
		return p.x, slot(p.y+1, p.y+p.h-2)
	case sideR:
		return p.x + p.w - 1, slot(p.y+1, p.y+p.h-2)
	case sideT:
		return slot(p.x+1, p.x+p.w-2), p.y
	default: // sideB
		return slot(p.x+1, p.x+p.w-2), p.y + p.h - 1
	}
}

// endpoint is a resolved connection: box side + on-canvas attach coordinate.
type endpoint struct {
	p       *placedEntity
	s       side
	x, y    int
	card    Cardinality
}

// drawConnectors routes every relationship and writes the result onto c.
func drawConnectors(c *canvas, lay *layout, d *ErDiagram, g glyphs) {
	o := newOverlay()

	// Decide each endpoint's side, then hand out attach slots per box-side so
	// multiple connectors on one face fan out instead of overlapping.
	type ends struct{ a, b endpoint }
	all := make([]ends, len(d.Relationships))
	slotCount := map[[2]int]int{} // (entityIdx, side) -> count
	entIdx := map[*placedEntity]int{}
	for i, p := range lay.placed {
		entIdx[p] = i
	}
	for i, r := range d.Relationships {
		a, b := lay.byName[r.Left], lay.byName[r.Right]
		if a == nil || b == nil {
			continue
		}
		sa, sb := sidesFor(a, b)
		all[i] = ends{
			a: endpoint{p: a, s: sa, card: r.LeftCard},
			b: endpoint{p: b, s: sb, card: r.RightCard},
		}
		slotCount[[2]int{entIdx[a], int(sa)}]++
		slotCount[[2]int{entIdx[b], int(sb)}]++
	}
	slotUsed := map[[2]int]int{}
	for i := range all {
		if all[i].a.p == nil {
			continue
		}
		for _, ep := range []*endpoint{&all[i].a, &all[i].b} {
			key := [2]int{entIdx[ep.p], int(ep.s)}
			ep.x, ep.y = attach(ep.p, ep.s, slotUsed[key], slotCount[key])
			slotUsed[key]++
		}
	}

	for i, r := range d.Relationships {
		e := all[i]
		if e.a.p == nil {
			continue
		}
		route(o, lay, e.a, e.b, r, i)
	}

	composite(c, o, g)
}

// sidesFor picks each box's exit face. Connectors leave through the left/right
// faces so the crow's-foot markers face the boxes naturally; same-column pairs
// both exit right into the shared gutter between that column and the next.
func sidesFor(a, b *placedEntity) (side, side) {
	switch {
	case b.col > a.col:
		return sideR, sideL
	case b.col < a.col:
		return sideL, sideR
	default: // same column (incl. self): both exit right
		return sideR, sideR
	}
}

// route draws one relationship as an orthogonal line in this edge's own lane, so
// distinct edges never overlap — only genuine crossings share a cell (┼).
// Neighbouring boxes get a straight (or single-jog) line across the gutter
// between them; boxes that aren't neighbours dip through a clear horizontal
// gutter so the run never cuts across an intervening box.
func route(o *overlay, lay *layout, a, b endpoint, r *Relationship, k int) {
	lane := k % lay.lanes

	if a.p == b.p { // self-relationship: a small loop off the right side
		selfLoop(o, lay, a, lane, r)
		return
	}

	putToken(o, a.x, a.y, a.s, r.LeftCard)
	putToken(o, b.x, b.y, b.s, r.RightCard)

	// Same column: join the two right sides through the gutter to their right.
	if a.p.col == b.p.col {
		xg := lay.vGutX[a.p.col+1] + lane
		o.polyline([][2]int{{a.x, a.y}, {xg, a.y}, {xg, b.y}, {b.x, b.y}}, r.Identifying)
		putLabel(o, r.Label, xg+1, xg+1+runewidth.StringWidth(r.Label), (a.y+b.y)/2)
		return
	}

	// Neighbouring columns: one shared gutter, straight line when rows align.
	if abs(a.p.col-b.p.col) == 1 {
		xg := lay.vGutX[maxI(a.p.col, b.p.col)] + lane
		o.polyline([][2]int{{a.x, a.y}, {xg, a.y}, {xg, b.y}, {b.x, b.y}}, r.Identifying)
		if a.y == b.y { // straight line: label spans the whole gap
			putLabel(o, r.Label, minI(a.x, b.x), maxI(a.x, b.x), a.y)
		} else if abs(xg-a.x) >= abs(b.x-xg) {
			putLabel(o, r.Label, minI(a.x, xg), maxI(a.x, xg), a.y)
		} else {
			putLabel(o, r.Label, minI(xg, b.x), maxI(xg, b.x), b.y)
		}
		return
	}

	// Distant columns: drop into a horizontal gutter and run across it, so the
	// line never crosses an intervening box.
	xa := lay.vGutX[a.p.col+1] + lane
	if a.s == sideL {
		xa = lay.vGutX[a.p.col] + lane
	}
	xb := lay.vGutX[b.p.col+1] + lane
	if b.s == sideL {
		xb = lay.vGutX[b.p.col] + lane
	}
	hg := maxI(a.p.row, b.p.row)
	if a.p.row == b.p.row {
		hg = a.p.row + 1
	}
	hy := lay.hGutY[hg] + lane
	o.polyline([][2]int{
		{a.x, a.y}, {xa, a.y}, {xa, hy}, {xb, hy}, {xb, b.y}, {b.x, b.y},
	}, r.Identifying)
	putLabel(o, r.Label, minI(xa, xb), maxI(xa, xb), hy)
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// selfLoop draws a compact loop from a box's right side back to itself.
func selfLoop(o *overlay, lay *layout, a endpoint, lane int, r *Relationship) {
	x := lay.vGutX[a.p.col+1] + lane
	y1, y2 := a.p.y+1, a.p.y+a.p.h-2
	ax := a.p.x + a.p.w - 1
	o.polyline([][2]int{{ax, y1}, {x, y1}, {x, y2}, {ax, y2}}, r.Identifying)
	putLabel(o, r.Label, ax, x, y1)
}

// putToken stamps a two-cell crow's-foot marker just outside a box on the stub,
// oriented so the marker reads toward the box. (atX,atY) is the box-edge attach
// cell and s the face the connector leaves from.
func putToken(o *overlay, atX, atY int, s side, card Cardinality) {
	var tok []rune
	var start int
	if s == sideR {
		tok = []rune(leftToken(card)) // }o, }|, |o … reading into the gutter
		start = atX + 1
	} else {
		tok = []rune(rightToken(card)) // o{, |{, o| … ending at the box
		start = atX - len(tok)
	}
	for i, r := range tok {
		o.token[[2]int{start + i, atY}] = r
	}
}

// putLabel centres a label on the run [x0,x1] at row y, clipped clear of the
// crow's-foot tokens (3 cells) at each end.
func putLabel(o *overlay, s string, x0, x1, y int) {
	if s == "" {
		return
	}
	x0 += 2
	x1 -= 2
	rs := []rune(s)
	start := (x0 + x1 - len(rs) + 1) / 2
	if start < x0 {
		start = x0
	}
	for i, r := range rs {
		if start+i > x1 {
			break
		}
		o.label[[2]int{start + i, y}] = r
	}
}

// composite renders the overlay onto the canvas: line junctions first (only on
// blank cells so boxes stay intact), then labels and crow's-foot tokens on top.
func composite(c *canvas, o *overlay, g glyphs) {
	seen := map[[2]int]bool{}
	mark := func(x, y int) {
		p := [2]int{x, y}
		if seen[p] {
			return
		}
		seen[p] = true
		bits := o.bits(x, y)
		if bits == 0 {
			return
		}
		if c.at(x, y) != ' ' {
			return // don't scribble over a box
		}
		c.set(x, y, glyphFor(bits, o.solid[p] != 0, g))
	}
	for p := range o.solid {
		mark(p[0], p[1])
	}
	for p := range o.dash {
		mark(p[0], p[1])
	}
	for p, r := range o.label {
		c.set(p[0], p[1], r)
	}
	for p, r := range o.token {
		c.set(p[0], p[1], r)
	}
}

// glyphFor maps a set of direction bits to a box-drawing rune.
func glyphFor(bits uint8, solid bool, g glyphs) rune {
	switch bits {
	case dN | dS:
		if solid {
			return g.v
		}
		return ':'
	case dE | dW:
		return g.h
	case dN | dE:
		return g.bl // └
	case dN | dW:
		return g.br // ┘
	case dS | dE:
		return g.tl // ┌
	case dS | dW:
		return g.tr // ┐
	case dN | dS | dE:
		return g.teeR // ├
	case dN | dS | dW:
		return g.teeL // ┤
	case dN | dE | dW:
		return g.teeU // ┴
	case dS | dE | dW:
		return g.teeD // ┬
	case dN | dS | dE | dW:
		return g.cross // ┼
	case dN, dS:
		if solid {
			return g.v
		}
		return ':'
	default: // dE, dW, 0
		return g.h
	}
}

// crow's-foot cardinality markers, read toward the box.
func leftToken(c Cardinality) string {
	switch c {
	case OnlyOne:
		return "||"
	case ZeroOrOne:
		return "|o"
	case ZeroOrMore:
		return "}o"
	default:
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
	default:
		return "|{"
	}
}

func sign(n int) int {
	if n > 0 {
		return 1
	}
	if n < 0 {
		return -1
	}
	return 0
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
