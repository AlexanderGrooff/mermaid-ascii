package sequence

import (
	"fmt"
	"strings"

	"github.com/AlexanderGrooff/mermaid-ascii/internal/diagram"
	"github.com/mattn/go-runewidth"
)

type sequenceFitPlan struct {
	participantSpacing  int
	selfMessageWidth    int
	labelPolicy         string
	participantLabelMax int
	messageLabelMax     int
	noteLabelMax        int
	blockLabelMax       int
}

func fitSequenceToWidth(sd *SequenceDiagram, config *diagram.Config) (string, error) {
	basePlan := sequenceFitPlan{
		participantSpacing:  config.SequenceParticipantSpacing,
		selfMessageWidth:    config.SequenceSelfMessageWidth,
		labelPolicy:         diagram.EdgeLabelPolicyFull,
		participantLabelMax: 0,
		messageLabelMax:     0,
		noteLabelMax:        0,
		blockLabelMax:       0,
	}

	plans := sequenceFitPlans(sd, basePlan, config.MaxWidth)
	bestOutput := ""
	bestWidth := 0
	for idx, plan := range plans {
		adjustedConfig := *config
		adjustedConfig.SequenceParticipantSpacing = plan.participantSpacing
		adjustedConfig.SequenceSelfMessageWidth = plan.selfMessageWidth

		adjustedDiagram := applySequenceFitPlan(sd, plan)
		output, err := renderSequenceBase(adjustedDiagram, &adjustedConfig)
		if err != nil {
			return "", err
		}
		width := maxOutputLineWidth(output)
		if idx == 0 || width < bestWidth {
			bestWidth = width
			bestOutput = output
		}
		if width <= config.MaxWidth {
			return output, nil
		}
	}

	return bestOutput, nil
}

func sequenceFitPlans(sd *SequenceDiagram, base sequenceFitPlan, maxWidth int) []sequenceFitPlan {
	plans := []sequenceFitPlan{}
	seen := map[string]struct{}{}
	addPlan := func(plan sequenceFitPlan) {
		normalized := normalizeSequencePlan(plan, maxWidth, sd)
		key := sequenceFitPlanKey(normalized)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		plans = append(plans, normalized)
	}

	addPlan(base)

	compact := base
	compact.participantSpacing = min(defaultParticipantSpacing, 2)
	compact.selfMessageWidth = min(defaultSelfMessageWidth, 3)
	addPlan(compact)

	tight := base
	tight.participantSpacing = 1
	tight.selfMessageWidth = 2
	addPlan(tight)

	ellipsis := base
	ellipsis.labelPolicy = diagram.EdgeLabelPolicyEllipsis
	addPlan(ellipsis)

	ellipsisCompact := compact
	ellipsisCompact.labelPolicy = diagram.EdgeLabelPolicyEllipsis
	addPlan(ellipsisCompact)

	ellipsisTight := tight
	ellipsisTight.labelPolicy = diagram.EdgeLabelPolicyEllipsis
	addPlan(ellipsisTight)

	drop := tight
	drop.labelPolicy = diagram.EdgeLabelPolicyDrop
	addPlan(drop)

	return plans
}

func normalizeSequencePlan(plan sequenceFitPlan, maxWidth int, sd *SequenceDiagram) sequenceFitPlan {
	if plan.participantSpacing <= 0 {
		plan.participantSpacing = defaultParticipantSpacing
	}
	if plan.selfMessageWidth <= 0 {
		plan.selfMessageWidth = defaultSelfMessageWidth
	}
	if plan.selfMessageWidth < 2 {
		plan.selfMessageWidth = 2
	}
	if plan.labelPolicy == "" {
		plan.labelPolicy = diagram.EdgeLabelPolicyFull
	}
	if plan.labelPolicy != diagram.EdgeLabelPolicyFull {
		plan.participantLabelMax = participantLabelMaxWidthFor(sd, plan, maxWidth)
		plan.messageLabelMax = messageLabelMaxWidthFor(maxWidth)
		plan.noteLabelMax = noteLabelMaxWidthFor(maxWidth)
		plan.blockLabelMax = blockLabelMaxWidthFor(maxWidth)
	}
	return plan
}

func participantLabelMaxWidthFor(sd *SequenceDiagram, plan sequenceFitPlan, maxWidth int) int {
	if maxWidth <= 0 || sd == nil || len(sd.Participants) == 0 {
		return 0
	}
	count := len(sd.Participants)
	totalSpacing := plan.participantSpacing * (count - 1)
	overhead := (boxPaddingLeftRight + boxBorderWidth) * count
	available := maxWidth - totalSpacing - overhead
	if available < count {
		return 1
	}
	return available / count
}

func messageLabelMaxWidthFor(maxWidth int) int {
	if maxWidth <= 0 {
		return 0
	}
	available := maxWidth - labelBufferSpace - labelLeftMargin - 4
	if available < 1 {
		return 1
	}
	return available
}

func noteLabelMaxWidthFor(maxWidth int) int {
	if maxWidth <= 0 {
		return 0
	}
	available := maxWidth - 4
	if available < 1 {
		return 1
	}
	return available
}

func blockLabelMaxWidthFor(maxWidth int) int {
	return noteLabelMaxWidthFor(maxWidth)
}

func sequenceFitPlanKey(plan sequenceFitPlan) string {
	return fmt.Sprintf("%d:%d:%s:%d:%d:%d:%d",
		plan.participantSpacing,
		plan.selfMessageWidth,
		plan.labelPolicy,
		plan.participantLabelMax,
		plan.messageLabelMax,
		plan.noteLabelMax,
		plan.blockLabelMax,
	)
}

func applySequenceFitPlan(sd *SequenceDiagram, plan sequenceFitPlan) *SequenceDiagram {
	if sd == nil {
		return nil
	}
	participants := make([]*Participant, len(sd.Participants))
	participantMap := make(map[*Participant]*Participant, len(sd.Participants))
	for i, p := range sd.Participants {
		label := applyLabelPolicy(p.Label, plan.labelPolicy, plan.participantLabelMax)
		cp := &Participant{
			ID:    p.ID,
			Label: label,
			Index: p.Index,
		}
		participants[i] = cp
		participantMap[p] = cp
	}

	var messages []*Message
	var cloneElement func(elem DiagramElement) DiagramElement
	cloneElement = func(elem DiagramElement) DiagramElement {
		switch e := elem.(type) {
		case *Message:
			label := applyLabelPolicy(e.Label, plan.labelPolicy, plan.messageLabelMax)
			msg := &Message{
				From:      participantMap[e.From],
				To:        participantMap[e.To],
				Label:     label,
				ArrowType: e.ArrowType,
				Number:    e.Number,
			}
			messages = append(messages, msg)
			return msg
		case *Note:
			text := applyLabelPolicy(e.Text, plan.labelPolicy, plan.noteLabelMax)
			actors := make([]*Participant, len(e.Actors))
			for i, actor := range e.Actors {
				actors[i] = participantMap[actor]
			}
			return &Note{
				Position: e.Position,
				Actors:   actors,
				Text:     text,
			}
		case *Block:
			return cloneBlock(e, plan, participantMap, cloneElement)
		default:
			return nil
		}
	}

	elements := make([]DiagramElement, len(sd.Elements))
	for i, elem := range sd.Elements {
		elements[i] = cloneElement(elem)
	}

	return &SequenceDiagram{
		Participants: participants,
		Messages:     messages,
		Elements:     elements,
		Autonumber:   sd.Autonumber,
	}
}

func cloneBlock(block *Block, plan sequenceFitPlan, participantMap map[*Participant]*Participant, cloneElement func(DiagramElement) DiagramElement) *Block {
	if block == nil {
		return nil
	}
	sections := make([]*BlockSection, len(block.Sections))
	for i, section := range block.Sections {
		label := applyLabelPolicy(section.Label, plan.labelPolicy, plan.blockLabelMax)
		sectionElements := make([]DiagramElement, len(section.Elements))
		for j, elem := range section.Elements {
			sectionElements[j] = cloneElement(elem)
		}
		sections[i] = &BlockSection{
			Label:    label,
			Elements: sectionElements,
		}
	}
	label := applyLabelPolicy(block.Label, plan.labelPolicy, plan.blockLabelMax)
	return &Block{
		Type:     block.Type,
		Label:    label,
		Sections: sections,
	}
}

func applyLabelPolicy(label, policy string, maxWidth int) string {
	if label == "" {
		return ""
	}
	switch policy {
	case diagram.EdgeLabelPolicyDrop:
		return ""
	case diagram.EdgeLabelPolicyEllipsis:
		return ellipsisLabel(label, maxWidth)
	default:
		return label
	}
}

func ellipsisLabel(label string, maxWidth int) string {
	if maxWidth <= 0 {
		return label
	}
	if runewidth.StringWidth(label) <= maxWidth {
		return label
	}
	if maxWidth <= 3 {
		return strings.Repeat(".", maxWidth)
	}
	trimmed := truncateToWidth(label, maxWidth-3)
	return trimmed + "..."
}

func truncateToWidth(label string, width int) string {
	if width <= 0 {
		return ""
	}
	var sb strings.Builder
	currentWidth := 0
	for _, r := range label {
		rw := runewidth.RuneWidth(r)
		if currentWidth+rw > width {
			break
		}
		sb.WriteRune(r)
		currentWidth += rw
	}
	return sb.String()
}

func maxOutputLineWidth(output string) int {
	lines := strings.Split(output, "\n")
	maxWidth := 0
	for _, line := range lines {
		width := runewidth.StringWidth(line)
		if width > maxWidth {
			maxWidth = width
		}
	}
	return maxWidth
}
