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

// widthCondition is used for calculating display width.
// EastAsianWidth is set to false so that box drawing characters
// are treated as narrow (width=1) while CJK characters remain wide (width=2).
var widthCondition = &runewidth.Condition{EastAsianWidth: false}

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
		w := widthCondition.StringWidth(p.Label) + boxPaddingLeftRight
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
		labelLen := widthCondition.StringWidth(sd.Participants[i].Label)
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

	for _, msg := range sd.Messages {
		for i := 0; i < layout.messageSpacing; i++ {
			lines = append(lines, buildLifeline(layout, chars))
		}

		if msg.From == msg.To {
			lines = append(lines, renderSelfMessage(msg, layout, chars)...)
		} else {
			lines = append(lines, renderMessage(msg, layout, chars)...)
		}
	}

	lines = append(lines, buildLifeline(layout, chars))
	return strings.Join(lines, "\n") + "\n", nil
}

func buildLine(participants []*Participant, layout *diagramLayout, draw func(int) string) string {
	var sb strings.Builder
	for i := range participants {
		boxWidth := layout.participantWidths[i] + boxBorderWidth
		left := layout.participantCenters[i] - boxWidth/2

		needed := left - widthCondition.StringWidth(sb.String())
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
		labelWidth := widthCondition.StringWidth(label)
		totalW := max(layout.totalWidth, start+labelWidth) + labelBufferSpace

		// Build the line using display width positioning
		var sb strings.Builder
		pos := 0 // current display position
		labelRunes := []rune(label)
		labelIdx := 0

		for pos < totalW {
			// Label takes priority over lifelines (overwrites them)
			if pos >= start && labelIdx < len(labelRunes) {
				r := labelRunes[labelIdx]
				sb.WriteRune(r)
				pos += widthCondition.RuneWidth(r)
				labelIdx++
			} else {
				// Check if current position is a lifeline
				isLifeline := false
				for _, c := range layout.participantCenters {
					if pos == c {
						isLifeline = true
						break
					}
				}

				if isLifeline {
					sb.WriteRune(chars.Vertical)
				} else {
					sb.WriteRune(' ')
				}
				pos++
			}
		}

		lines = append(lines, strings.TrimRight(sb.String(), " "))
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
		labelWidth := widthCondition.StringWidth(label)
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
