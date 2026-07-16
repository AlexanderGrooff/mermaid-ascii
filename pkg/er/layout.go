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
// Runes advance by display width: a double-width rune (CJK, emoji) occupies
// its cell plus a sentinel cell, keeping canvas columns aligned with what the
// terminal shows.
func (c *canvas) stamp(x0, y0 int, block []string) {
	for dy, line := range block {
		x := x0
		for _, r := range line {
			c.set(x, y0+dy, r)
			w := runewidth.RuneWidth(r)
			if w == 2 {
				c.set(x+1, y0+dy, 0)
			}
			x += w
		}
	}
}

func (c *canvas) String() string {
	var b strings.Builder
	for _, row := range c.rows {
		line := make([]rune, 0, len(row))
		for _, r := range row {
			if r != 0 { // sentinel: second column of a double-width rune
				line = append(line, r)
			}
		}
		b.WriteString(strings.TrimRight(string(line), " "))
		b.WriteByte('\n')
	}
	return b.String()
}

// side identifies which face of a box a connector attaches to. Connectors only
// ever leave through the top or bottom face (see sidesFor).
type side int

const (
	sideT side = iota
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

	// Vertical gutters hold one trunk lane per relationship AND must leave the
	// widest label room on a horizontal run. A relationship's longer run always
	// spans at least gutW-lane cells, so budgeting lanes+label keeps every
	// label whole whichever lane it lands in.
	maxLabel := 0
	for _, r := range d.Relationships {
		if w := runewidth.StringWidth(r.Label); w > maxLabel {
			maxLabel = w
		}
	}
	gutW := lanes + maxLabel + 5

	// Boxes must be wide enough that every relationship touching them gets its
	// own attach column (self-loops additionally need room for two crow's-foot
	// tokens plus a gap between their stubs).
	deg := map[string]int{}
	selfLoop := map[string]bool{}
	for _, r := range d.Relationships {
		deg[r.Left]++
		deg[r.Right]++
		if r.Left == r.Right {
			selfLoop[r.Left] = true
		}
	}

	placed := make([]*placedEntity, n)
	for i, e := range d.Entities {
		minW := 2*deg[e.Name] + 3
		if selfLoop[e.Name] {
			minW = max(minW, 9)
		}
		lines := renderEntity(e, g, minW-2)
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
	return &layout{byName: byName, placed: placed, lanes: lanes, gutW: gutW, vGutX: vGutX, hGutY: hGutY}
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

// attach chooses evenly-spaced connection points along a box's top or bottom
// face so several relationships touching the same face don't stack on one
// cell. It returns the (x,y) for the idx-th of total connectors on that face.
// placeEntities sizes boxes so span ≥ 2·total, keeping every slot distinct.
func attach(p *placedEntity, s side, idx, total int) (int, int) {
	lo, hi := p.x+1, p.x+p.w-2
	span := hi - lo
	x := lo + span/2
	if total > 1 && span > 0 {
		x = lo + span*(idx+1)/(total+1)
	}
	y := p.y + p.h - 1 // sideB
	if s == sideT {
		y = p.y
	}
	return x, y
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

	plans := make([]routePlan, 0, len(d.Relationships))
	for i, r := range d.Relationships {
		if all[i].a.p == nil {
			continue
		}
		plans = append(plans, newPlan(lay, all[i].a, all[i].b, r, i))
	}

	// Lines first, decorations second: label placement inspects the finished
	// line overlay so labels can dodge cells other relationships pass through.
	for _, p := range plans {
		p.drawLine(o)
	}
	for _, p := range plans {
		p.decorate(o)
	}

	composite(c, o, g)

	// Mark each attach point on the box border with a tee so the stub visibly
	// joins the box (composite itself never draws over box cells).
	for _, p := range plans {
		setAttachTee(c, p.a, g)
		setAttachTee(c, p.b, g)
	}
}

// setAttachTee stamps ┬/┴ where a stub leaves a box; if the border cell already
// tees the other way (an attribute-table column rule), the two merge into ┼.
func setAttachTee(c *canvas, ep endpoint, g glyphs) {
	tee, opposite := g.teeD, g.teeU
	if ep.s == sideT {
		tee, opposite = g.teeU, g.teeD
	}
	if c.at(ep.x, ep.y) == opposite {
		tee = g.cross
	}
	c.set(ep.x, ep.y, tee)
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

// routePlan is one relationship's resolved geometry. Each endpoint drops (or
// rises) vertically from its box into a horizontal gutter row; if both stubs
// reach the same row the run merges into one straight span, otherwise a
// vertical trunk in a side gutter joins the two rows:
//
//	{a.x,a.y} {a.x,ya} {tx,ya} {tx,yb} {b.x,yb} {b.x,b.y}
type routePlan struct {
	rel    *Relationship
	a, b   endpoint
	ya, yb int
	tx     int  // trunk column, valid only when !merged
	merged bool // both stubs meet one gutter row: single run, no trunk
}

func newPlan(lay *layout, a, b endpoint, r *Relationship, lane int) routePlan {
	ya, yb := lay.gutterY(a, lane), lay.gutterY(b, lane)
	p := routePlan{rel: r, a: a, b: b, ya: ya, yb: yb, merged: ya == yb}
	if !p.merged {
		p.tx = lay.trunkX(a.p, b.p, lane)
	}
	return p
}

// drawLine draws the relationship's orthogonal line in its own lane, so
// distinct edges never overlap — only genuine crossings share a cell (┼).
func (p routePlan) drawLine(o *overlay) {
	if p.merged {
		o.polyline([][2]int{
			{p.a.x, p.a.y}, {p.a.x, p.ya}, {p.b.x, p.ya}, {p.b.x, p.b.y},
		}, p.rel.Identifying)
		return
	}
	o.polyline([][2]int{
		{p.a.x, p.a.y}, {p.a.x, p.ya}, {p.tx, p.ya}, {p.tx, p.yb}, {p.b.x, p.yb}, {p.b.x, p.b.y},
	}, p.rel.Identifying)
}

// decorate stamps the crow's-foot tokens and the label. Runs after every line
// is drawn so the label can dodge cells other relationships pass through.
func (p routePlan) decorate(o *overlay) {
	if p.merged {
		putToken(o, p.a, p.b.x, p.ya)
		putToken(o, p.b, p.a.x, p.ya)
		if p.a.p == p.b.p { // self-loop: the run is at most the box's width, so
			// the label sits beside the loop instead of inside it
			writeLabel(o, p.rel.Label, max(p.a.x, p.b.x)+2, p.ya, -1)
		} else {
			putLabel(o, p.rel.Label, [][3]int{{min(p.a.x, p.b.x), max(p.a.x, p.b.x), p.ya}})
		}
		return
	}
	putToken(o, p.a, p.tx, p.ya)
	putToken(o, p.b, p.tx, p.yb)
	runs := [][3]int{
		{min(p.a.x, p.tx), max(p.a.x, p.tx), p.ya},
		{min(p.b.x, p.tx), max(p.b.x, p.tx), p.yb},
	}
	if runs[1][1]-runs[1][0] > runs[0][1]-runs[0][0] {
		runs[0], runs[1] = runs[1], runs[0] // longest run first
	}
	putLabel(o, p.rel.Label, runs)
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

// putLabel places a label on one of the candidate runs (each {x0,x1,y}),
// clipped 3 cells at each end to clear the corner and crow's-foot token.
// Candidates arrive longest-first; the first one the label fits on wins.
func putLabel(o *overlay, s string, runs [][3]int) {
	if s == "" {
		return
	}
	lw := runewidth.StringWidth(s)
	best := runs[0]
	for _, r := range runs {
		if r[1]-r[0]-6+1 >= lw {
			best = r
			break
		}
	}
	lo, hi, y := best[0]+3, best[1]-3, best[2]
	if lo > hi {
		return
	}
	writeLabel(o, s, labelStart(o, lo, hi, lw, y), y, hi)
}

// labelStart picks where the label begins: centred on the run, sliding
// outwards if needed to a spot where it severs no crossing vertical line.
func labelStart(o *overlay, lo, hi, lw, y int) int {
	centre := max(lo, lo+(hi-lo+1-lw)/2)
	for d := 0; d <= hi-lo; d++ {
		for _, c := range []int{centre - d, centre + d} {
			if c >= lo && c+lw-1 <= hi && !vBlocked(o, c, c+lw-1, y) {
				return c
			}
		}
	}
	return centre
}

// vBlocked reports whether any cell in [x0,x1] at row y carries a vertical
// line (a crossing another relationship's trunk or stub makes through here).
func vBlocked(o *overlay, x0, x1, y int) bool {
	for x := x0; x <= x1; x++ {
		if o.bits(x, y)&(dN|dS) != 0 {
			return true
		}
	}
	return false
}

// writeLabel writes label runes from x, advancing by display width (reserving
// a sentinel cell after each double-width rune so columns stay aligned). A
// limit ≥ 0 clips the label; -1 writes it whole.
func writeLabel(o *overlay, s string, x, y, limit int) {
	for _, c := range s {
		w := runewidth.RuneWidth(c)
		if limit >= 0 && x+w-1 > limit {
			return
		}
		o.label[[2]int{x, y}] = c
		if w == 2 {
			o.label[[2]int{x + 1, y}] = 0
		}
		x += w
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

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
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
