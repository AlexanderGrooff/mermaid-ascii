package sequence

import (
	"fmt"
	"strings"

	"github.com/AlexanderGrooff/mermaid-ascii/pkg/diagram"
	"github.com/mattn/go-runewidth"
)

const (
	defaultSelfMessageWidth   = 4
	defaultMessageSpacing     = 1
	defaultParticipantSpacing = 5
	boxPaddingLeftRight       = 2
	minBoxWidth               = 3
	boxBorderWidth            = 2
	labelLeftMargin           = 2
	labelBufferSpace          = 10
	frameIndent               = 2 // columns reserved per nested fragment level
	frameLabelInset           = 2 // columns from the left corner to the "[label]" tab
)

type diagramLayout struct {
	participantWidths  []int
	participantCenters []int
	totalWidth         int
	messageSpacing     int
	selfMessageWidth   int
}

func calculateLayout(sd *SequenceDiagram, config *diagram.Config) *diagramLayout {
	participantSpacing := config.SequenceParticipantSpacing
	if participantSpacing <= 0 {
		participantSpacing = defaultParticipantSpacing
	}

	widths := make([]int, len(sd.Participants))
	for i, p := range sd.Participants {
		w := runewidth.StringWidth(p.Label) + boxPaddingLeftRight
		if w < minBoxWidth {
			w = minBoxWidth
		}
		widths[i] = w
	}

	centers := make([]int, len(sd.Participants))
	currentX := 0
	for i := range sd.Participants {
		boxWidth := widths[i] + boxBorderWidth
		if i == 0 {
			centers[i] = boxWidth / 2
			currentX = boxWidth
		} else {
			currentX += participantSpacing
			centers[i] = currentX + boxWidth/2
			currentX += boxWidth
		}
	}

	last := len(sd.Participants) - 1
	totalWidth := centers[last] + (widths[last]+boxBorderWidth)/2

	msgSpacing := config.SequenceMessageSpacing
	if msgSpacing <= 0 {
		msgSpacing = defaultMessageSpacing
	}
	selfWidth := config.SequenceSelfMessageWidth
	if selfWidth <= 0 {
		selfWidth = defaultSelfMessageWidth
	}

	return &diagramLayout{
		participantWidths:  widths,
		participantCenters: centers,
		totalWidth:         totalWidth,
		messageSpacing:     msgSpacing,
		selfMessageWidth:   selfWidth,
	}
}

func Render(sd *SequenceDiagram, config *diagram.Config) (string, error) {
	if sd == nil || len(sd.Participants) == 0 {
		return "", fmt.Errorf("no participants")
	}
	if config == nil {
		config = diagram.DefaultConfig()
	}

	chars := Unicode
	if config.UseAscii {
		chars = ASCII
	}

	layout := calculateLayout(sd, config)

	// Fall back to a message-only body for diagrams built without an event
	// stream (e.g. constructed by hand rather than via Parse).
	events := sd.Events
	if len(events) == 0 {
		for _, msg := range sd.Messages {
			events = append(events, Event{Kind: EventMessage, Message: msg})
		}
	}

	// Nested fragment frames stack their left borders in a gutter to the left of
	// the first participant. A single (unnested) fragment already fits with its
	// border at column 0, so only levels beyond the first need reserved columns;
	// shift the whole diagram right so the borders never overlap the lifelines.
	if gutter := (fragmentDepth(events) - 1) * frameIndent; gutter > 0 {
		shiftLayoutRight(layout, gutter)
	}

	// A "left of" note on the leftmost participant, or a wide "over" note,
	// extends past column 0. Reserve an additional left gutter so no note box is
	// clamped on top of lifelines it shouldn't cover.
	if gutter := noteLeftGutter(events, layout); gutter > 0 {
		shiftLayoutRight(layout, gutter)
	}

	var lines []string

	lines = append(lines, buildLine(sd.Participants, layout, func(i int) string {
		return string(chars.TopLeft) + strings.Repeat(string(chars.Horizontal), layout.participantWidths[i]) + string(chars.TopRight)
	}))

	lines = append(lines, buildLine(sd.Participants, layout, func(i int) string {
		w := layout.participantWidths[i]
		labelLen := runewidth.StringWidth(sd.Participants[i].Label)
		pad := (w - labelLen) / 2
		return string(chars.Vertical) + strings.Repeat(" ", pad) + sd.Participants[i].Label +
			strings.Repeat(" ", w-pad-labelLen) + string(chars.Vertical)
	}))

	lines = append(lines, buildLine(sd.Participants, layout, func(i int) string {
		w := layout.participantWidths[i]
		return string(chars.BottomLeft) + strings.Repeat(string(chars.Horizontal), w/2) +
			string(chars.TeeDown) + strings.Repeat(string(chars.Horizontal), w-w/2-1) +
			string(chars.BottomRight)
	}))

	lines = append(lines, renderEvents(events, layout, chars)...)

	lines = append(lines, buildLifeline(layout, chars))
	return strings.Join(lines, "\n") + "\n", nil
}

// renderEvents paints the ordered body of the diagram — messages and fragment
// frames — into text lines. It recurses into each loop/opt block, so nested
// fragments (a loop containing an opt, say) render correctly.
func renderEvents(events []Event, layout *diagramLayout, chars BoxChars) []string {
	var lines []string
	for i := 0; i < len(events); {
		ev := events[i]
		if ev.Kind == EventFragmentStart {
			end := matchingFragmentEnd(events, i)
			lines = append(lines, wrapFragment(ev.Fragment, events[i+1:end], layout, chars)...)
			i = end + 1
			continue
		}

		// A lone EventFragmentEnd is normally consumed by matchingFragmentEnd,
		// so it only reaches here if given an unbalanced event stream; skip it
		// defensively rather than nil-deref on ev.Message below.
		if ev.Kind == EventFragmentEnd {
			i++
			continue
		}

		if ev.Kind == EventNote {
			for s := 0; s < layout.messageSpacing; s++ {
				lines = append(lines, buildLifeline(layout, chars))
			}
			lines = append(lines, renderNote(ev.Note, layout, chars)...)
			i++
			continue
		}

		// EventMessage.
		msg := ev.Message
		for s := 0; s < layout.messageSpacing; s++ {
			lines = append(lines, buildLifeline(layout, chars))
		}
		if msg.From == msg.To {
			lines = append(lines, renderSelfMessage(msg, layout, chars)...)
		} else {
			lines = append(lines, renderMessage(msg, layout, chars)...)
		}
		i++
	}
	return lines
}

// fragmentDepth returns the maximum fragment nesting depth within events (0 if
// there are no fragments).
func fragmentDepth(events []Event) int {
	maxDepth, cur := 0, 0
	for _, ev := range events {
		switch ev.Kind {
		case EventFragmentStart:
			cur++
			if cur > maxDepth {
				maxDepth = cur
			}
		case EventFragmentEnd:
			cur--
		}
	}
	return maxDepth
}

// matchingFragmentEnd returns the index of the EventFragmentEnd that closes the
// fragment opened at start, accounting for nested fragments.
func matchingFragmentEnd(events []Event, start int) int {
	depth := 0
	for i := start; i < len(events); i++ {
		switch events[i].Kind {
		case EventFragmentStart:
			depth++
		case EventFragmentEnd:
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return len(events) // unreachable: the parser guarantees balanced fragments
}

// renderNote draws a note annotation as a bordered box positioned over or
// beside its participant lifelines. The box obscures any lifelines it covers,
// while lifelines outside it stay continuous.
// shiftLayoutRight moves every participant (and the total width) right by n
// columns, reserving a left gutter for content that extends past column 0.
func shiftLayoutRight(layout *diagramLayout, n int) {
	for i := range layout.participantCenters {
		layout.participantCenters[i] += n
	}
	layout.totalWidth += n
}

// noteLeftGutter returns how many columns the diagram must shift right so that
// every note box — and the border of each fragment frame enclosing it — fits at
// or right of column 0. A note nested d fragments deep needs room for its box
// plus one border column per enclosing frame plus a one-column margin.
func noteLeftGutter(events []Event, layout *diagramLayout) int {
	gutter, depth := 0, 0
	for _, ev := range events {
		switch ev.Kind {
		case EventFragmentStart:
			depth++
		case EventFragmentEnd:
			depth--
		case EventNote:
			left, _ := noteBoxColumns(ev.Note, layout)
			// A top-level note only needs its own box at column 0. A note d
			// frames deep also needs the outermost enclosing border, which sits
			// 1+(d-1)*frameIndent columns to its left (see wrapFragment's stagger).
			extra := 0
			if depth > 0 {
				extra = 1 + (depth-1)*frameIndent
			}
			if need := -left + extra; need > gutter {
				gutter = need
			}
		}
	}
	return gutter
}

// noteRunes returns a note's display text with mermaid line breaks collapsed to
// spaces (ASCII output is single-line).
func noteRunes(note *Note) []rune {
	text := note.Text
	for _, br := range []string{"<br/>", "<br />", "<br>"} {
		text = strings.ReplaceAll(text, br, " ")
	}
	return []rune(text)
}

// noteBoxColumns returns the [left, right] columns a note's box occupies. For
// `over`, the box always spans its participants (first..last) and widens
// symmetrically to fit the text; for left/right of it sits beside the lifeline.
// left may be negative when a left-of box extends past column 0 — Render
// reserves a gutter so that never happens at draw time.
func noteBoxColumns(note *Note, layout *diagramLayout) (int, int) {
	boxW := len(noteRunes(note)) + 4 // "│ text │"
	centers := layout.participantCenters
	first := centers[note.Participants[0].Index]
	last := centers[note.Participants[len(note.Participants)-1].Index]

	switch note.Placement {
	case NoteRightOf:
		left := first + 2
		return left, left + boxW - 1
	case NoteLeftOf:
		right := first - 2
		return right - boxW + 1, right
	default: // NoteOver: span the named participants, widen for text, keep centred
		lo, hi := first, last
		if hi < lo {
			lo, hi = hi, lo
		}
		left, right := lo-1, hi+1
		if extra := boxW - (right - left + 1); extra > 0 {
			left -= extra / 2
			right += extra - extra/2
		}
		return left, right
	}
}

// renderNote draws a note annotation as a bordered box positioned over or
// beside its participant lifelines. The box obscures any lifelines it covers,
// while lifelines outside it stay continuous.
func renderNote(note *Note, layout *diagramLayout, chars BoxChars) []string {
	runes := noteRunes(note)
	left, right := noteBoxColumns(note, layout)
	if left < 0 { // safety; Render's note gutter should already prevent this
		left = 0
	}

	border := func(l, r rune) string {
		line := padRunes(buildLifeline(layout, chars), right+1)
		line[left] = l
		for c := left + 1; c < right; c++ {
			line[c] = chars.Horizontal
		}
		line[right] = r
		return strings.TrimRight(string(line), " ")
	}

	mid := padRunes(buildLifeline(layout, chars), right+1)
	for c := left; c <= right; c++ { // clear covered lifelines
		mid[c] = ' '
	}
	mid[left] = chars.Vertical
	mid[right] = chars.Vertical
	// Centre the text within the box interior [left+1, right-1].
	inner := right - left - 1
	col := left + 1 + (inner-len(runes))/2
	for _, ch := range runes {
		if col > left && col < right {
			mid[col] = ch
		}
		col++
	}

	return []string{
		border(chars.TopLeft, chars.TopRight),
		strings.TrimRight(string(mid), " "),
		border(chars.BottomLeft, chars.BottomRight),
	}
}

// wrapFragment renders a loop/opt block: it paints the inner body, then draws a
// labelled frame around the participants the block touches.
func wrapFragment(frag *Fragment, inner []Event, layout *diagramLayout, chars BoxChars) []string {
	// Render the inner body first. A trailing lifeline gives breathing room
	// above the bottom border.
	body := renderEvents(inner, layout, chars)
	body = append(body, buildLifeline(layout, chars))

	// The frame spans from just left of the leftmost involved lifeline to just
	// right of the rightmost — the same participants the block's messages touch.
	// Its left border is pushed further left by the depth of frames nested
	// inside it, so each enclosing frame sits outside its children (the gutter
	// reserved in Render guarantees there is room).
	leftIdx, rightIdx := involvedParticipants(inner, layout)
	leftCol := layout.participantCenters[leftIdx] - frameIndent*(fragmentDepth(inner)+1)
	rightCol := layout.participantCenters[rightIdx] + frameIndent

	// A note in the body can extend beyond the participant span (a "left of"
	// note, or one wider than its span). Widen this frame to contain any note
	// box so its border never cuts through it. The margin scales with the note's
	// depth *relative to this frame* (rd) so that when several frames enclose the
	// same note, each outer frame lands frameIndent columns further out than the
	// one inside it (rather than all collapsing onto the note's edge).
	rd := 0
	for _, ev := range inner {
		switch ev.Kind {
		case EventFragmentStart:
			rd++
		case EventFragmentEnd:
			rd--
		case EventNote:
			nl, nr := noteBoxColumns(ev.Note, layout)
			if x := nl - 1 - rd*frameIndent; x < leftCol {
				leftCol = x
			}
			if x := nr + 1 + rd*frameIndent; x > rightCol {
				rightCol = x
			}
		}
	}
	if leftCol < 0 {
		leftCol = 0
	}

	// Message labels can extend well past the rightmost lifeline, so widen the
	// frame to clear the longest inner line.
	for _, l := range body {
		if w := len([]rune(l)) + 1; w > rightCol {
			rightCol = w
		}
	}

	label := frag.Type.String()
	if frag.Label != "" {
		label += " " + frag.Label
	}

	// The label tab ("[label]") sits frameLabelInset cells in from the left
	// corner; make sure the frame is wide enough to hold it without truncation.
	if labelEnd := leftCol + frameLabelInset + len([]rune("["+label+"]")) + 1; labelEnd > rightCol {
		rightCol = labelEnd
	}

	out := []string{fragmentBorder(layout, chars, leftCol, rightCol, label, true)}
	for _, l := range body {
		out = append(out, overlayFrameSides(l, chars, leftCol, rightCol))
	}
	out = append(out, fragmentBorder(layout, chars, leftCol, rightCol, "", false))
	return out
}

// involvedParticipants returns the smallest and largest participant indices
// referenced by the messages in events. If there are no messages it falls back
// to spanning every participant.
func involvedParticipants(events []Event, layout *diagramLayout) (int, int) {
	minIdx, maxIdx := -1, -1
	note := func(idx int) {
		if minIdx == -1 || idx < minIdx {
			minIdx = idx
		}
		if maxIdx == -1 || idx > maxIdx {
			maxIdx = idx
		}
	}
	for _, ev := range events {
		if ev.Kind == EventMessage {
			note(ev.Message.From.Index)
			note(ev.Message.To.Index)
		}
	}
	if minIdx == -1 {
		return 0, len(layout.participantCenters) - 1
	}
	return minIdx, maxIdx
}

// fragmentBorder builds a top or bottom frame border on top of a lifeline row,
// so participant lines outside the frame stay continuous. When top is true and
// label is non-empty, the label is embedded as a "[label]" tab near the left
// corner.
func fragmentBorder(layout *diagramLayout, chars BoxChars, leftCol, rightCol int, label string, top bool) string {
	line := padRunes(buildLifeline(layout, chars), rightCol+1)

	leftCorner, rightCorner := chars.BottomLeft, chars.BottomRight
	if top {
		leftCorner, rightCorner = chars.TopLeft, chars.TopRight
	}
	line[leftCol] = leftCorner
	for c := leftCol + 1; c < rightCol; c++ {
		line[c] = chars.Horizontal
	}
	line[rightCol] = rightCorner

	if label != "" {
		col := leftCol + frameLabelInset
		for _, r := range "[" + label + "]" {
			if col < rightCol {
				line[col] = r
				col++
			}
		}
	}
	return strings.TrimRight(string(line), " ")
}

// overlayFrameSides draws the left and right vertical borders of a frame onto an
// already-rendered content line.
func overlayFrameSides(line string, chars BoxChars, leftCol, rightCol int) string {
	r := padRunes(line, rightCol+1)
	r[leftCol] = chars.Vertical
	r[rightCol] = chars.Vertical
	return strings.TrimRight(string(r), " ")
}

// padRunes returns s as a rune slice right-padded with spaces to at least width.
func padRunes(s string, width int) []rune {
	r := []rune(s)
	for len(r) < width {
		r = append(r, ' ')
	}
	return r
}

func buildLine(participants []*Participant, layout *diagramLayout, draw func(int) string) string {
	var sb strings.Builder
	for i := range participants {
		boxWidth := layout.participantWidths[i] + boxBorderWidth
		left := layout.participantCenters[i] - boxWidth/2

		needed := left - len([]rune(sb.String()))
		if needed > 0 {
			sb.WriteString(strings.Repeat(" ", needed))
		}
		sb.WriteString(draw(i))
	}
	return sb.String()
}

func buildLifeline(layout *diagramLayout, chars BoxChars) string {
	line := make([]rune, layout.totalWidth+1)
	for i := range line {
		line[i] = ' '
	}
	for _, c := range layout.participantCenters {
		if c < len(line) {
			line[c] = chars.Vertical
		}
	}
	return strings.TrimRight(string(line), " ")
}

func renderMessage(msg *Message, layout *diagramLayout, chars BoxChars) []string {
	var lines []string
	from, to := layout.participantCenters[msg.From.Index], layout.participantCenters[msg.To.Index]

	label := msg.Label
	if msg.Number > 0 {
		label = fmt.Sprintf("%d. %s", msg.Number, msg.Label)
	}

	if label != "" {
		start := min(from, to) + labelLeftMargin
		labelWidth := runewidth.StringWidth(label)
		w := max(layout.totalWidth, start+labelWidth) + labelBufferSpace
		line := []rune(buildLifeline(layout, chars))
		if len(line) < w {
			padding := make([]rune, w-len(line))
			for k := range padding {
				padding[k] = ' '
			}
			line = append(line, padding...)
		}

		col := start
		for _, r := range label {
			if col < len(line) {
				line[col] = r
				col++
			}
		}
		lines = append(lines, strings.TrimRight(string(line), " "))
	}

	line := []rune(buildLifeline(layout, chars))
	style := chars.SolidLine
	if msg.ArrowType.isDotted() {
		style = chars.DottedLine
	}

	if from < to {
		line[from] = chars.TeeRight
		for i := from + 1; i < to; i++ {
			line[i] = style
		}
		// Open arrows (-> / -->) have no head: draw the line right up to the
		// target lifeline instead of an arrowhead.
		line[to-1] = style
		if msg.ArrowType.hasHead() {
			line[to-1] = chars.ArrowRight
		}
		line[to] = chars.Vertical
	} else {
		line[to] = chars.Vertical
		line[to+1] = style
		if msg.ArrowType.hasHead() {
			line[to+1] = chars.ArrowLeft
		}
		for i := to + 2; i < from; i++ {
			line[i] = style
		}
		line[from] = chars.TeeLeft
	}
	lines = append(lines, strings.TrimRight(string(line), " "))
	return lines
}

func renderSelfMessage(msg *Message, layout *diagramLayout, chars BoxChars) []string {
	var lines []string
	center := layout.participantCenters[msg.From.Index]
	width := layout.selfMessageWidth

	ensureWidth := func(l string) []rune {
		target := layout.totalWidth + width + 1
		r := []rune(l)
		if len(r) < target {
			pad := make([]rune, target-len(r))
			for i := range pad {
				pad[i] = ' '
			}
			r = append(r, pad...)
		}
		return r
	}

	label := msg.Label
	if msg.Number > 0 {
		label = fmt.Sprintf("%d. %s", msg.Number, msg.Label)
	}

	if label != "" {
		line := ensureWidth(buildLifeline(layout, chars))
		start := center + labelLeftMargin
		labelWidth := runewidth.StringWidth(label)
		needed := start + labelWidth + labelBufferSpace
		if len(line) < needed {
			pad := make([]rune, needed-len(line))
			for i := range pad {
				pad[i] = ' '
			}
			line = append(line, pad...)
		}
		col := start
		for _, c := range label {
			if col < len(line) {
				line[col] = c
				col++
			}
		}
		lines = append(lines, strings.TrimRight(string(line), " "))
	}

	// Solid arrows keep the solid horizontal glyph; dotted arrows (-->>/-->) use
	// the dotted line. For solid arrows style == chars.Horizontal, so output is
	// unchanged from before.
	style := chars.Horizontal
	if msg.ArrowType.isDotted() {
		style = chars.DottedLine
	}

	l1 := ensureWidth(buildLifeline(layout, chars))
	l1[center] = chars.TeeRight
	for i := 1; i < width; i++ {
		l1[center+i] = style
	}
	l1[center+width-1] = chars.SelfTopRight
	lines = append(lines, strings.TrimRight(string(l1), " "))

	l2 := ensureWidth(buildLifeline(layout, chars))
	l2[center+width-1] = chars.Vertical
	lines = append(lines, strings.TrimRight(string(l2), " "))

	l3 := ensureWidth(buildLifeline(layout, chars))
	l3[center] = chars.Vertical
	// Open arrows have no head.
	l3[center+1] = style
	if msg.ArrowType.hasHead() {
		l3[center+1] = chars.ArrowLeft
	}
	for i := 2; i < width-1; i++ {
		l3[center+i] = style
	}
	l3[center+width-1] = chars.SelfBottom
	lines = append(lines, strings.TrimRight(string(l3), " "))

	return lines
}
