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
		for i := range layout.participantCenters {
			layout.participantCenters[i] += gutter
		}
		layout.totalWidth += gutter
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

		// EventMessage. (EventFragmentEnd is always consumed as the boundary
		// found by matchingFragmentEnd, so it never reaches this branch.)
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
	if leftCol < 0 {
		leftCol = 0
	}
	rightCol := layout.participantCenters[rightIdx] + 2

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

	// The label tab ("[label]") sits two cells in from the left corner; make
	// sure the frame is wide enough to hold it without truncation.
	if labelEnd := leftCol + 2 + len([]rune("["+label+"]")) + 1; labelEnd > rightCol {
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
		col := leftCol + 2
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
	if msg.ArrowType == DottedArrow {
		style = chars.DottedLine
	}

	if from < to {
		line[from] = chars.TeeRight
		for i := from + 1; i < to; i++ {
			line[i] = style
		}
		line[to-1] = chars.ArrowRight
		line[to] = chars.Vertical
	} else {
		line[to] = chars.Vertical
		line[to+1] = chars.ArrowLeft
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

	l1 := ensureWidth(buildLifeline(layout, chars))
	l1[center] = chars.TeeRight
	for i := 1; i < width; i++ {
		l1[center+i] = chars.Horizontal
	}
	l1[center+width-1] = chars.SelfTopRight
	lines = append(lines, strings.TrimRight(string(l1), " "))

	l2 := ensureWidth(buildLifeline(layout, chars))
	l2[center+width-1] = chars.Vertical
	lines = append(lines, strings.TrimRight(string(l2), " "))

	l3 := ensureWidth(buildLifeline(layout, chars))
	l3[center] = chars.Vertical
	l3[center+1] = chars.ArrowLeft
	for i := 2; i < width-1; i++ {
		l3[center+i] = chars.Horizontal
	}
	l3[center+width-1] = chars.SelfBottom
	lines = append(lines, strings.TrimRight(string(l3), " "))

	return lines
}
