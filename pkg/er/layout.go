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
	gutW   int   // width of each vertical gutter
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

	// X layout: [col0][gutter][col1]…[gutter]. No gutter left of column 0 —
	// trunks only ever run between columns or off the right edge.
	vGutX := make([]int, cols+1)
	colX := make([]int, cols)
	x := 0
	for c := 0; c < cols; c++ {
		vGutX[c] = x
		if c > 0 {
			x += gutW
			colX[c] = x
		} else {
			colX[0] = 0
		}
		x += colW[c]
	}
	vGutX[cols] = x

	// Y layout: [row0][gutter][row1]…[gutter]. No gutter above row 0 — nothing
	// ever exits upward from the top row, so it would only be blank space.
	hGutY := make([]int, rows+1)
	rowY := make([]int, rows)
	y := 0
	for r := 0; r < rows; r++ {
		hGutY[r] = y
		if r > 0 {
			y += lanes
			rowY[r] = y
		} else {
			rowY[0] = 0
		}
		y += rowH[r]
	}
	hGutY[rows] = y

	byName := map[string]*placedEntity{}
	for _, p := range placed {
		p.x, p.y = colX[p.col], rowY[p.row]
		byName[p.entity.Name] = p
	}
	return &layout{byName: byName, placed: placed, lanes: lanes, gutW: gutW, vGutX: vGutX, hGutY: hGutY, cols: cols}
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
		for x != b[0] || y != b[1] {
			// outgoing link on the cell we leave — never on the segment's end,
			// or corners would grow phantom arms and render as tees.
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
			x += dx
			y += dy
			// incoming link on the cell we enter
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
	p    *placedEntity
	s    side
	x, y int
	card Cardinality
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
		// A self-loop's two stubs share one face; evenly-spaced slots sit too
		// close together for the crow's-foot tokens, so spread them to the ends.
		if p := all[i].a.p; p == all[i].b.p {
			all[i].a.x, all[i].b.x = p.x+1, p.x+p.w-2
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

	// Mark each attach point on the box border with a tee so the stub visibly
	// joins the box (composite itself never draws over box cells).
	for i := range all {
		if all[i].a.p == nil {
			continue
		}
		for _, ep := range []endpoint{all[i].a, all[i].b} {
			tee := g.teeD // ┬: stub leaves downward through a bottom border
			if ep.s == sideT {
				tee = g.teeU // ┴: stub leaves upward through a top border
			}
			c.set(ep.x, ep.y, tee)
		}
	}
}

// sidesFor picks each box's exit face. Connectors leave through the top/bottom
// faces so every relationship rides a horizontal gutter row — the only rows
// where labels are guaranteed collision-free (each relationship owns a global
// lane). Boxes exit toward each other, with one exception: vertically-adjacent
// boxes in the same column would meet in their shared gutter as a straight
// vertical line with no horizontal run to carry tokens or a label, so the lower
// box exits bottom too and the path wraps round through the side gutter.
func sidesFor(a, b *placedEntity) (side, side) {
	sameColAdjacent := a.col == b.col && abs(a.row-b.row) == 1
	switch {
	case a.row < b.row:
		if sameColAdjacent {
			return sideB, sideB
		}
		return sideB, sideT
	case a.row > b.row:
		if sameColAdjacent {
			return sideB, sideB
		}
		return sideT, sideB
	default: // same row (incl. self-relationships): both dip into the gutter below
		return sideB, sideB
	}
}

// gutterY is the horizontal gutter row an endpoint's stub reaches: the gutter
// below the box for bottom exits, above for top exits, offset by the
// relationship's own lane so distinct relationships never share a row.
func (l *layout) gutterY(e endpoint, lane int) int {
	if e.s == sideB {
		return l.hGutY[e.p.row+1] + lane
	}
	return l.hGutY[e.p.row] + lane
}

// trunkX is the vertical lane column a relationship travels along when its two
// gutter rows differ: a box-free vertical gutter between (or beside) the two
// columns, offset by the relationship's lane.
func (l *layout) trunkX(a, b *placedEntity, lane int) int {
	if a.col == b.col {
		// Far side of the gutter, so the horizontal runs span its full width
		// and have room to carry the relationship label.
		return l.vGutX[a.col+1] + l.gutW - 1 - lane
	}
	return l.vGutX[(a.col+b.col+1)/2] + lane
}

// route draws one relationship as an orthogonal line in this edge's own lane, so
// distinct edges never overlap — only genuine crossings share a cell (┼). Each
// endpoint drops (or rises) vertically from its box into a horizontal gutter
// row; if both stubs reach the same row the run merges into one straight span,
// otherwise a vertical trunk in a side gutter joins the two rows:
//
//	{a.x,a.y} {a.x,ya} {tx,ya} {tx,yb} {b.x,yb} {b.x,b.y}
func route(o *overlay, lay *layout, a, b endpoint, r *Relationship, k int) {
	lane := k % lay.lanes
	ya, yb := lay.gutterY(a, lane), lay.gutterY(b, lane)

	if ya == yb { // both stubs meet one gutter row: a single horizontal run
		o.polyline([][2]int{{a.x, a.y}, {a.x, ya}, {b.x, ya}, {b.x, b.y}}, r.Identifying)
		putToken(o, a, b.x, ya)
		putToken(o, b, a.x, ya)
		if a.p == b.p { // self-loop: the run is at most the box's width, so the
			// label sits beside the loop instead of inside it
			for i, c := range []rune(r.Label) {
				o.label[[2]int{maxI(a.x, b.x) + 2 + i, ya}] = c
			}
		} else {
			putLabel(o, r.Label, minI(a.x, b.x), maxI(a.x, b.x), ya)
		}
		return
	}

	tx := lay.trunkX(a.p, b.p, lane)
	o.polyline([][2]int{
		{a.x, a.y}, {a.x, ya}, {tx, ya}, {tx, yb}, {b.x, yb}, {b.x, b.y},
	}, r.Identifying)
	putToken(o, a, tx, ya)
	putToken(o, b, tx, yb)
	if abs(a.x-tx) >= abs(b.x-tx) { // label rides the longer horizontal run
		putLabel(o, r.Label, minI(a.x, tx), maxI(a.x, tx), ya)
	} else {
		putLabel(o, r.Label, minI(b.x, tx), maxI(b.x, tx), yb)
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// putToken stamps a two-cell crow's-foot marker on the gutter row, next to the
// corner where the endpoint's stub turns toward targetX. Orientation is pure
// geometry: whichever way the run leaves the stub, the marker's foot character
// stays adjacent to the box and the pair reads left-to-right.
func putToken(o *overlay, ep endpoint, targetX, y int) {
	if ep.x < targetX { // run leaves rightward
		for i, r := range []rune(leftToken(ep.card)) {
			o.token[[2]int{ep.x + 1 + i, y}] = r
		}
		return
	}
	tok := []rune(rightToken(ep.card))
	for i, r := range tok {
		o.token[[2]int{ep.x - len(tok) + i, y}] = r
	}
}

// putLabel centres a label on the run [x0,x1] at row y, clipped clear of the
// corner cell plus two-cell crow's-foot token (3 cells) at each end.
func putLabel(o *overlay, s string, x0, x1, y int) {
	if s == "" {
		return
	}
	x0 += 3
	x1 -= 3
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
		return g.vd
	case dE | dW:
		if solid {
			return g.h
		}
		return g.hd
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
		return g.vd
	default: // dE, dW, 0
		if solid {
			return g.h
		}
		return g.hd
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
