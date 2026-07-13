package er

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

// box-drawing glyphs (Unicode by default, ASCII when useAscii).
type glyphs struct {
	h, v, tl, tr, bl, br, teeD, teeU, teeL, teeR, cross rune
}

var unicodeGlyphs = glyphs{'─', '│', '┌', '┐', '└', '┘', '┬', '┴', '┤', '├', '┼'}
var asciiGlyphs = glyphs{'-', '|', '+', '+', '+', '+', '+', '+', '+', '+', '+'}

// renderEntity draws an entity as an attribute table: a name header above a grid
// of the attribute rows. Columns (type, name, key, comment) are included only
// when at least one attribute uses them, and are padded to a common width.
func renderEntity(e *Entity, g glyphs) []string {
	// Column cells per attribute row: type, name, keys, comment.
	rows := make([][4]string, len(e.Attributes))
	has := [4]bool{true, true, false, false} // type/name always shown
	for i, a := range e.Attributes {
		rows[i] = [4]string{a.Type, a.Name, strings.Join(a.Keys, ","), a.Comment}
		if rows[i][2] != "" {
			has[2] = true
		}
		if rows[i][3] != "" {
			has[3] = true
		}
	}

	// Which columns are shown, and each shown column's width.
	var cols []int // indices of shown columns
	for c := 0; c < 4; c++ {
		if has[c] {
			cols = append(cols, c)
		}
	}
	width := map[int]int{}
	for _, c := range cols {
		for _, r := range rows {
			if w := runewidth.StringWidth(r[c]); w > width[c] {
				width[c] = w
			}
		}
	}

	// Inner width = sum of padded cells (" cell ") + separators between columns.
	inner := 0
	for i, c := range cols {
		inner += width[c] + 2 // one space padding each side
		if i > 0 {
			inner++ // column separator
		}
	}
	// The header (entity name) may be wider than the columns; grow the last
	// column so the grid and header line up.
	if nameW := runewidth.StringWidth(e.Name) + 2; nameW > inner && len(cols) > 0 {
		width[cols[len(cols)-1]] += nameW - inner
		inner = nameW
	}

	pad := func(s string, w int) string {
		return " " + s + strings.Repeat(" ", w-runewidth.StringWidth(s)) + " "
	}
	rule := func(left, mid, right rune) string {
		var b strings.Builder
		b.WriteRune(left)
		for i, c := range cols {
			if i > 0 {
				b.WriteRune(mid)
			}
			b.WriteString(strings.Repeat(string(g.h), width[c]+2))
		}
		b.WriteRune(right)
		return b.String()
	}

	var out []string
	// Top border + centred name header + separator with column tees.
	out = append(out, string(g.tl)+strings.Repeat(string(g.h), inner)+string(g.tr))
	namePad := inner - runewidth.StringWidth(e.Name)
	out = append(out, string(g.v)+strings.Repeat(" ", namePad/2)+e.Name+
		strings.Repeat(" ", namePad-namePad/2)+string(g.v))
	out = append(out, rule(g.teeR, g.teeD, g.teeL))
	// Attribute rows.
	for _, r := range rows {
		var b strings.Builder
		b.WriteRune(g.v)
		for i, c := range cols {
			if i > 0 {
				b.WriteRune(g.v)
			}
			b.WriteString(pad(r[c], width[c]))
		}
		b.WriteRune(g.v)
		out = append(out, b.String())
	}
	out = append(out, rule(g.bl, g.teeU, g.br))
	return out
}
