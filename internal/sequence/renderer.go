package sequence

import (
	"fmt"
	"strings"

	"github.com/AlexanderGrooff/mermaid-ascii/internal/diagram"
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

	for _, elem := range sd.Elements {
		for i := 0; i < layout.messageSpacing; i++ {
			lines = append(lines, buildLifeline(layout, chars))
		}

		switch e := elem.(type) {
		case *Message:
			if e.From == e.To {
				lines = append(lines, renderSelfMessage(e, layout, chars)...)
			} else {
				lines = append(lines, renderMessage(e, layout, chars)...)
			}
		case *Note:
			noteLines := renderNote(e, layout, chars)
			if noteLines != nil {
				lines = append(lines, noteLines...)
			}
		case *Block:
			blockLines := renderBlock(e, layout, chars, 0, layout.messageSpacing)
			if blockLines != nil {
				lines = append(lines, blockLines...)
			}
		}
	}

	lines = append(lines, buildLifeline(layout, chars))
	return strings.Join(lines, "\n") + "\n", nil
}

func buildLine(participants []*Participant, layout *diagramLayout, draw func(int) string) string {
	var sb strings.Builder
	currentWidth := 0
	for i := range participants {
		boxWidth := layout.participantWidths[i] + boxBorderWidth
		left := layout.participantCenters[i] - boxWidth/2

		needed := left - currentWidth
		if needed > 0 {
			sb.WriteString(strings.Repeat(" ", needed))
			currentWidth += needed
		}
		content := draw(i)
		sb.WriteString(content)
		currentWidth += runewidth.StringWidth(content)
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

func renderNote(note *Note, layout *diagramLayout, chars BoxChars) []string {
	switch note.Position {
	case NoteOver:
		return renderNoteOver(note, layout, chars)
	case NoteLeftOf:
		return renderNoteLeftOf(note, layout, chars)
	case NoteRightOf:
		return renderNoteRightOf(note, layout, chars)
	}
	return nil
}

func renderNoteOver(note *Note, layout *diagramLayout, chars BoxChars) []string {
	var lines []string

	leftActor := note.Actors[0]
	rightActor := note.Actors[len(note.Actors)-1]

	leftCenter := layout.participantCenters[leftActor.Index]
	rightCenter := layout.participantCenters[rightActor.Index]

	if leftCenter > rightCenter {
		leftCenter, rightCenter = rightCenter, leftCenter
	}

	padding := 2
	textWidth := runewidth.StringWidth(note.Text)
	minBoxWidth := textWidth + 4
	spanWidth := rightCenter - leftCenter + padding*2
	boxWidth := spanWidth
	if boxWidth < minBoxWidth {
		boxWidth = minBoxWidth
	}

	spanCenter := (leftCenter + rightCenter) / 2
	boxLeft := spanCenter - boxWidth/2
	if boxLeft < 0 {
		boxLeft = 0
	}
	boxRight := boxLeft + boxWidth

	topLine := make([]rune, layout.totalWidth+boxWidth)
	for i := range topLine {
		topLine[i] = ' '
	}
	for _, c := range layout.participantCenters {
		if c < len(topLine) {
			if c >= boxLeft && c <= boxRight {
				topLine[c] = chars.TeeUp
			} else {
				topLine[c] = chars.Vertical
			}
		}
	}
	topLine[boxLeft] = chars.TopLeft
	for i := boxLeft + 1; i < boxRight; i++ {
		if topLine[i] != chars.TeeUp {
			topLine[i] = chars.Horizontal
		}
	}
	topLine[boxRight] = chars.TopRight
	lines = append(lines, strings.TrimRight(string(topLine), " "))

	textLine := make([]rune, layout.totalWidth+boxWidth)
	for i := range textLine {
		textLine[i] = ' '
	}
	for _, c := range layout.participantCenters {
		if c < len(textLine) && (c < boxLeft || c > boxRight) {
			textLine[c] = chars.Vertical
		}
	}
	textLine[boxLeft] = chars.Vertical
	textLine[boxRight] = chars.Vertical
	textStart := boxLeft + (boxWidth-textWidth)/2
	col := textStart
	for _, r := range note.Text {
		if col < len(textLine) && col < boxRight {
			textLine[col] = r
			col++
		}
	}
	lines = append(lines, strings.TrimRight(string(textLine), " "))

	bottomLine := make([]rune, layout.totalWidth+boxWidth)
	for i := range bottomLine {
		bottomLine[i] = ' '
	}
	for _, c := range layout.participantCenters {
		if c < len(bottomLine) {
			if c >= boxLeft && c <= boxRight {
				bottomLine[c] = chars.TeeDown
			} else {
				bottomLine[c] = chars.Vertical
			}
		}
	}
	bottomLine[boxLeft] = chars.BottomLeft
	for i := boxLeft + 1; i < boxRight; i++ {
		if bottomLine[i] != chars.TeeDown {
			bottomLine[i] = chars.Horizontal
		}
	}
	bottomLine[boxRight] = chars.BottomRight
	lines = append(lines, strings.TrimRight(string(bottomLine), " "))

	return lines
}

func renderNoteLeftOf(note *Note, layout *diagramLayout, chars BoxChars) []string {
	var lines []string
	actor := note.Actors[0]
	center := layout.participantCenters[actor.Index]

	textWidth := runewidth.StringWidth(note.Text)
	boxWidth := textWidth + 4
	boxRight := center - 2
	boxLeft := boxRight - boxWidth

	if boxLeft < 0 {
		boxLeft = 0
		boxRight = boxWidth
	}

	ensureWidth := layout.totalWidth + 1
	if boxRight >= ensureWidth {
		ensureWidth = boxRight + 1
	}

	topLine := make([]rune, ensureWidth)
	for i := range topLine {
		topLine[i] = ' '
	}
	for _, c := range layout.participantCenters {
		if c < len(topLine) && (c < boxLeft || c > boxRight) {
			topLine[c] = chars.Vertical
		}
	}
	topLine[boxLeft] = chars.TopLeft
	for i := boxLeft + 1; i < boxRight; i++ {
		topLine[i] = chars.Horizontal
	}
	topLine[boxRight] = chars.TopRight
	lines = append(lines, strings.TrimRight(string(topLine), " "))

	textLine := make([]rune, ensureWidth)
	for i := range textLine {
		textLine[i] = ' '
	}
	for _, c := range layout.participantCenters {
		if c < len(textLine) && (c < boxLeft || c > boxRight) {
			textLine[c] = chars.Vertical
		}
	}
	textLine[boxLeft] = chars.Vertical
	if boxRight < center {
		textLine[boxRight] = chars.TeeLeft
		for i := boxRight + 1; i < center; i++ {
			textLine[i] = chars.Horizontal
		}
		textLine[center] = chars.TeeLeft
	} else {
		textLine[boxRight] = chars.Vertical
	}
	textStart := boxLeft + 2
	col := textStart
	for _, r := range note.Text {
		if col < boxRight {
			textLine[col] = r
			col++
		}
	}
	lines = append(lines, strings.TrimRight(string(textLine), " "))

	bottomLine := make([]rune, ensureWidth)
	for i := range bottomLine {
		bottomLine[i] = ' '
	}
	for _, c := range layout.participantCenters {
		if c < len(bottomLine) && (c < boxLeft || c > boxRight) {
			bottomLine[c] = chars.Vertical
		}
	}
	bottomLine[boxLeft] = chars.BottomLeft
	for i := boxLeft + 1; i < boxRight; i++ {
		bottomLine[i] = chars.Horizontal
	}
	bottomLine[boxRight] = chars.BottomRight
	lines = append(lines, strings.TrimRight(string(bottomLine), " "))

	return lines
}

func renderNoteRightOf(note *Note, layout *diagramLayout, chars BoxChars) []string {
	var lines []string
	actor := note.Actors[0]
	center := layout.participantCenters[actor.Index]

	textWidth := runewidth.StringWidth(note.Text)
	boxWidth := textWidth + 4
	boxLeft := center + 2
	boxRight := boxLeft + boxWidth

	ensureWidth := layout.totalWidth
	if boxRight >= ensureWidth {
		ensureWidth = boxRight + 1
	}

	// Top border
	topLine := make([]rune, ensureWidth)
	for i := range topLine {
		topLine[i] = ' '
	}
	for _, c := range layout.participantCenters {
		if c < len(topLine) {
			topLine[c] = chars.Vertical
		}
	}
	topLine[boxLeft] = chars.TopLeft
	for i := boxLeft + 1; i < boxRight; i++ {
		topLine[i] = chars.Horizontal
	}
	topLine[boxRight] = chars.TopRight
	lines = append(lines, strings.TrimRight(string(topLine), " "))

	// Text line with connector
	textLine := make([]rune, ensureWidth)
	for i := range textLine {
		textLine[i] = ' '
	}
	for _, c := range layout.participantCenters {
		if c < len(textLine) {
			textLine[c] = chars.Vertical
		}
	}
	textLine[center] = chars.TeeRight
	for i := center + 1; i < boxLeft; i++ {
		textLine[i] = chars.Horizontal
	}
	textLine[boxLeft] = chars.TeeRight
	textLine[boxRight] = chars.Vertical
	// Add text
	textStart := boxLeft + 2
	col := textStart
	for _, r := range note.Text {
		if col < boxRight {
			textLine[col] = r
			col++
		}
	}
	lines = append(lines, strings.TrimRight(string(textLine), " "))

	// Bottom border
	bottomLine := make([]rune, ensureWidth)
	for i := range bottomLine {
		bottomLine[i] = ' '
	}
	for _, c := range layout.participantCenters {
		if c < len(bottomLine) {
			bottomLine[c] = chars.Vertical
		}
	}
	bottomLine[boxLeft] = chars.BottomLeft
	for i := boxLeft + 1; i < boxRight; i++ {
		bottomLine[i] = chars.Horizontal
	}
	bottomLine[boxRight] = chars.BottomRight
	lines = append(lines, strings.TrimRight(string(bottomLine), " "))

	return lines
}

// calculateBlockRightEdge returns the minimum rightmost column index needed to
// contain all content in the block. This accounts for message labels, note texts,
// section labels, and nested blocks.
func calculateBlockRightEdge(block *Block, layout *diagramLayout) int {
	maxRightEdge := 0

	var checkElements func(elements []DiagramElement)
	checkElements = func(elements []DiagramElement) {
		for _, elem := range elements {
			switch e := elem.(type) {
			case *Message:
				label := e.Label
				if e.Number > 0 {
					label = fmt.Sprintf("%d. %s", e.Number, e.Label)
				}
				if label != "" {
					from := layout.participantCenters[e.From.Index]
					to := layout.participantCenters[e.To.Index]
					start := min(from, to) + labelLeftMargin
					labelWidth := runewidth.StringWidth(label)
					rightEdge := start + labelWidth + 2
					if rightEdge > maxRightEdge {
						maxRightEdge = rightEdge
					}
				}
			case *Note:
				textWidth := runewidth.StringWidth(e.Text)
				var noteRight int
				switch e.Position {
				case NoteOver:
					leftCenter := layout.participantCenters[e.Actors[0].Index]
					rightCenter := layout.participantCenters[e.Actors[len(e.Actors)-1].Index]
					if leftCenter > rightCenter {
						leftCenter, rightCenter = rightCenter, leftCenter
					}
					spanCenter := (leftCenter + rightCenter) / 2
					minBoxWidth := textWidth + 4
					spanWidth := rightCenter - leftCenter + 4
					boxWidth := spanWidth
					if boxWidth < minBoxWidth {
						boxWidth = minBoxWidth
					}
					noteRight = spanCenter + boxWidth/2 + 1
				case NoteRightOf:
					center := layout.participantCenters[e.Actors[0].Index]
					noteRight = center + 2 + textWidth + 4
				case NoteLeftOf:
					noteRight = layout.participantCenters[e.Actors[0].Index]
				}
				if noteRight > maxRightEdge {
					maxRightEdge = noteRight
				}
			case *Block:
				nestedRightEdge := calculateBlockRightEdge(e, layout)
				if nestedRightEdge > maxRightEdge {
					maxRightEdge = nestedRightEdge
				}
			}
		}
	}

	for _, section := range block.Sections {
		sectionLabelWidth := runewidth.StringWidth(section.Label)
		if sectionLabelWidth+4 > maxRightEdge {
			maxRightEdge = sectionLabelWidth + 4
		}
		checkElements(section.Elements)
	}

	return maxRightEdge
}

func findBlockParticipantRange(block *Block) (minIdx, maxIdx int) {
	minIdx = -1
	maxIdx = -1

	var findInElements func(elements []DiagramElement)
	findInElements = func(elements []DiagramElement) {
		for _, elem := range elements {
			switch e := elem.(type) {
			case *Message:
				if minIdx == -1 || e.From.Index < minIdx {
					minIdx = e.From.Index
				}
				if minIdx == -1 || e.To.Index < minIdx {
					minIdx = e.To.Index
				}
				if e.From.Index > maxIdx {
					maxIdx = e.From.Index
				}
				if e.To.Index > maxIdx {
					maxIdx = e.To.Index
				}
			case *Note:
				for _, actor := range e.Actors {
					if minIdx == -1 || actor.Index < minIdx {
						minIdx = actor.Index
					}
					if actor.Index > maxIdx {
						maxIdx = actor.Index
					}
				}
			case *Block:
				nestedMin, nestedMax := findBlockParticipantRange(e)
				if nestedMin != -1 {
					if minIdx == -1 || nestedMin < minIdx {
						minIdx = nestedMin
					}
					if nestedMax > maxIdx {
						maxIdx = nestedMax
					}
				}
			}
		}
	}

	for _, section := range block.Sections {
		findInElements(section.Elements)
	}

	return minIdx, maxIdx
}

func renderBlock(block *Block, layout *diagramLayout, chars BoxChars, depth int, messageSpacing int) []string {
	var lines []string

	minIdx, maxIdx := findBlockParticipantRange(block)
	if minIdx == -1 || maxIdx == -1 {
		minIdx = 0
		maxIdx = len(layout.participantCenters) - 1
	}

	indent := depth * 2
	leftCenter := layout.participantCenters[minIdx]
	rightCenter := layout.participantCenters[maxIdx]

	boxLeft := leftCenter - 3 + indent
	if boxLeft < 0 {
		boxLeft = 0
	}
	boxRight := rightCenter + 3 - indent

	headerLabel := fmt.Sprintf("%s %s", block.Type, block.Label)
	labelWidth := runewidth.StringWidth(headerLabel)
	if boxRight-boxLeft < labelWidth+4 {
		boxRight = boxLeft + labelWidth + 4
	}

	contentRightEdge := calculateBlockRightEdge(block, layout)
	if contentRightEdge > boxRight {
		boxRight = contentRightEdge
	}

	ensureWidth := boxRight + 1
	if ensureWidth < layout.totalWidth {
		ensureWidth = layout.totalWidth
	}

	bc := GetBlockChars(block.Type, chars)

	makeLine := func() []rune {
		line := make([]rune, ensureWidth+1)
		for i := range line {
			line[i] = ' '
		}
		for _, c := range layout.participantCenters {
			if c < len(line) {
				line[c] = chars.Vertical
			}
		}
		return line
	}

	topLine := makeLine()
	topLine[boxLeft] = bc.TopLeft
	for i := boxLeft + 1; i < boxRight; i++ {
		if topLine[i] == chars.Vertical {
			topLine[i] = bc.TeeUp
		} else {
			topLine[i] = bc.Horizontal
		}
	}
	topLine[boxRight] = bc.TopRight
	lines = append(lines, strings.TrimRight(string(topLine), " "))

	headerLine := makeLine()
	headerLine[boxLeft] = bc.Vertical
	headerLine[boxRight] = bc.Vertical
	col := boxLeft + 2
	for _, r := range headerLabel {
		if col < boxRight {
			headerLine[col] = r
			col++
		}
	}
	lines = append(lines, strings.TrimRight(string(headerLine), " "))

	sepLine := makeLine()
	sepLine[boxLeft] = bc.TeeRight
	for i := boxLeft + 1; i < boxRight; i++ {
		if sepLine[i] == chars.Vertical {
			sepLine[i] = bc.Cross
		} else {
			sepLine[i] = bc.Horizontal
		}
	}
	sepLine[boxRight] = bc.TeeLeft
	lines = append(lines, strings.TrimRight(string(sepLine), " "))

	for sectionIdx, section := range block.Sections {
		if sectionIdx > 0 {
			divLine := makeLine()
			divLine[boxLeft] = bc.TeeRight
			for i := boxLeft + 1; i < boxRight; i++ {
				if divLine[i] == chars.Vertical {
					divLine[i] = bc.Cross
				} else {
					divLine[i] = bc.Horizontal
				}
			}
			divLine[boxRight] = bc.TeeLeft
			lines = append(lines, strings.TrimRight(string(divLine), " "))

			if section.Label != "" {
				labelLine := makeLine()
				labelLine[boxLeft] = bc.Vertical
				labelLine[boxRight] = bc.Vertical
				col := boxLeft + 2
				for _, r := range section.Label {
					if col < boxRight {
						labelLine[col] = r
						col++
					}
				}
				lines = append(lines, strings.TrimRight(string(labelLine), " "))
			}
		}

		for _, elem := range section.Elements {
			for i := 0; i < messageSpacing; i++ {
				spaceLine := makeLine()
				spaceLine[boxLeft] = bc.Vertical
				spaceLine[boxRight] = bc.Vertical
				lines = append(lines, strings.TrimRight(string(spaceLine), " "))
			}

			switch e := elem.(type) {
			case *Message:
				msgLines := renderMessage(e, layout, chars)
				for _, ml := range msgLines {
					mlRunes := []rune(ml)
					for len(mlRunes) <= ensureWidth {
						mlRunes = append(mlRunes, ' ')
					}
					mlRunes[boxLeft] = bc.Vertical
					mlRunes[boxRight] = bc.Vertical
					lines = append(lines, strings.TrimRight(string(mlRunes), " "))
				}
			case *Note:
				noteLines := renderNote(e, layout, chars)
				for _, nl := range noteLines {
					nlRunes := []rune(nl)
					for len(nlRunes) <= ensureWidth {
						nlRunes = append(nlRunes, ' ')
					}
					nlRunes[boxLeft] = bc.Vertical
					nlRunes[boxRight] = bc.Vertical
					lines = append(lines, strings.TrimRight(string(nlRunes), " "))
				}
			case *Block:
				nestedLines := renderBlock(e, layout, chars, depth+1, messageSpacing)
				for _, nl := range nestedLines {
					nlRunes := []rune(nl)
					for len(nlRunes) <= ensureWidth {
						nlRunes = append(nlRunes, ' ')
					}
					nlRunes[boxLeft] = bc.Vertical
					nlRunes[boxRight] = bc.Vertical
					lines = append(lines, strings.TrimRight(string(nlRunes), " "))
				}
			}
		}
	}

	spaceLine := makeLine()
	spaceLine[boxLeft] = bc.Vertical
	spaceLine[boxRight] = bc.Vertical
	lines = append(lines, strings.TrimRight(string(spaceLine), " "))

	bottomLine := makeLine()
	bottomLine[boxLeft] = bc.BottomLeft
	for i := boxLeft + 1; i < boxRight; i++ {
		if bottomLine[i] == chars.Vertical {
			bottomLine[i] = bc.TeeDown
		} else {
			bottomLine[i] = bc.Horizontal
		}
	}
	bottomLine[boxRight] = bc.BottomRight
	lines = append(lines, strings.TrimRight(string(bottomLine), " "))

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
