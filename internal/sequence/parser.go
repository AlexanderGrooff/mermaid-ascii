package sequence

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/AlexanderGrooff/mermaid-ascii/internal/diagram"
)

const (
	SequenceDiagramKeyword = "sequenceDiagram"
	SolidArrowSyntax       = "->>"
	DottedArrowSyntax      = "-->>"
)

var (
	// participantRegex matches participant declarations: participant [ID] [as Label]
	participantRegex = regexp.MustCompile(`^\s*participant\s+(?:"([^"]+)"|(\S+))(?:\s+as\s+(.+))?$`)

	// messageRegex matches messages: [From]->>[To]: [Label]
	messageRegex = regexp.MustCompile(`^\s*(?:"([^"]+)"|([^\s\->]+))\s*(-->>|->>)\s*(?:"([^"]+)"|([^\s\->]+))\s*:\s*(.*)$`)

	// autonumberRegex matches the autonumber directive
	autonumberRegex = regexp.MustCompile(`^\s*autonumber\s*$`)

	// noteRegex matches note declarations:
	//   Note over Actor: text
	//   Note over Actor1,Actor2: text
	//   Note left of Actor: text
	//   Note right of Actor: text
	noteRegex = regexp.MustCompile(`(?i)^\s*note\s+(over|left\s+of|right\s+of)\s+([^:]+):\s*(.*)$`)

	// blockStartRegex matches block start: loop, alt, opt, par, critical, break, rect
	blockStartRegex = regexp.MustCompile(`(?i)^\s*(loop|alt|opt|par|critical|break|rect)\s*(.*)$`)

	// blockDividerRegex matches block dividers: else, and, option
	blockDividerRegex = regexp.MustCompile(`(?i)^\s*(else|and|option)\s*(.*)$`)

	// blockEndRegex matches block end
	blockEndRegex = regexp.MustCompile(`(?i)^\s*end\s*$`)
)

// SequenceDiagram represents a parsed sequence diagram.
type SequenceDiagram struct {
	Participants []*Participant
	Messages     []*Message
	Elements     []DiagramElement // Ordered messages and notes
	Autonumber   bool
}

type Participant struct {
	ID    string
	Label string
	Index int
}

type Message struct {
	From      *Participant
	To        *Participant
	Label     string
	ArrowType ArrowType
	Number    int // Message number when autonumber is enabled (0 means no number)
}

type ArrowType int

const (
	SolidArrow ArrowType = iota
	DottedArrow
)

func (a ArrowType) String() string {
	switch a {
	case SolidArrow:
		return "solid"
	case DottedArrow:
		return "dotted"
	default:
		return fmt.Sprintf("ArrowType(%d)", a)
	}
}

type NotePosition int

const (
	NoteOver NotePosition = iota
	NoteLeftOf
	NoteRightOf
)

func (n NotePosition) String() string {
	switch n {
	case NoteOver:
		return "over"
	case NoteLeftOf:
		return "left of"
	case NoteRightOf:
		return "right of"
	default:
		return fmt.Sprintf("NotePosition(%d)", n)
	}
}

type Note struct {
	Position NotePosition
	Actors   []*Participant
	Text     string
}

type BlockType int

const (
	BlockLoop BlockType = iota
	BlockAlt
	BlockOpt
	BlockPar
	BlockCritical
	BlockBreak
	BlockRect
)

func (b BlockType) String() string {
	switch b {
	case BlockLoop:
		return "loop"
	case BlockAlt:
		return "alt"
	case BlockOpt:
		return "opt"
	case BlockPar:
		return "par"
	case BlockCritical:
		return "critical"
	case BlockBreak:
		return "break"
	case BlockRect:
		return "rect"
	default:
		return fmt.Sprintf("BlockType(%d)", b)
	}
}

type BlockSection struct {
	Label    string
	Elements []DiagramElement
}

type Block struct {
	Type     BlockType
	Label    string
	Sections []*BlockSection
}

func (*Block) isElement() {}

type DiagramElement interface {
	isElement()
}

func parseBlockType(keyword string) (BlockType, error) {
	switch strings.ToLower(keyword) {
	case "loop":
		return BlockLoop, nil
	case "alt":
		return BlockAlt, nil
	case "opt":
		return BlockOpt, nil
	case "par":
		return BlockPar, nil
	case "critical":
		return BlockCritical, nil
	case "break":
		return BlockBreak, nil
	case "rect":
		return BlockRect, nil
	default:
		return 0, fmt.Errorf("unknown block type: %q", keyword)
	}
}

func isValidDivider(blockType BlockType, divider string) bool {
	divider = strings.ToLower(divider)
	switch blockType {
	case BlockAlt:
		return divider == "else"
	case BlockPar:
		return divider == "and"
	case BlockCritical:
		return divider == "option"
	default:
		return false
	}
}

func (*Message) isElement() {}
func (*Note) isElement()    {}

func IsSequenceDiagram(input string) bool {
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}
		return strings.HasPrefix(trimmed, SequenceDiagramKeyword)
	}
	return false
}

func Parse(input string) (*SequenceDiagram, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty input")
	}

	rawLines := diagram.SplitLines(input)
	lines := diagram.RemoveComments(rawLines)
	if len(lines) == 0 {
		return nil, fmt.Errorf("no content found")
	}

	if !strings.HasPrefix(strings.TrimSpace(lines[0]), SequenceDiagramKeyword) {
		return nil, fmt.Errorf("expected %q keyword", SequenceDiagramKeyword)
	}
	lines = lines[1:]

	sd := &SequenceDiagram{
		Participants: []*Participant{},
		Messages:     []*Message{},
		Autonumber:   false,
	}
	participantMap := make(map[string]*Participant)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Check for autonumber directive
		if autonumberRegex.MatchString(trimmed) {
			sd.Autonumber = true
			continue
		}

		if matched, err := sd.parseParticipant(trimmed, participantMap); err != nil {
			return nil, fmt.Errorf("line %d: %w", i+2, err)
		} else if matched {
			continue
		}

		if matched, err := sd.parseMessage(trimmed, participantMap); err != nil {
			return nil, fmt.Errorf("line %d: %w", i+2, err)
		} else if matched {
			continue
		}

		if matched, err := sd.parseNote(trimmed, participantMap); err != nil {
			return nil, fmt.Errorf("line %d: %w", i+2, err)
		} else if matched {
			continue
		}

		return nil, fmt.Errorf("line %d: invalid syntax: %q", i+2, trimmed)
	}

	if len(sd.Participants) == 0 {
		return nil, fmt.Errorf("no participants found")
	}

	return sd, nil
}

func (sd *SequenceDiagram) parseParticipant(line string, participants map[string]*Participant) (bool, error) {
	match := participantRegex.FindStringSubmatch(line)
	if match == nil {
		return false, nil
	}

	id := match[2]
	if match[1] != "" {
		id = match[1]
	}
	label := match[3]
	if label == "" {
		label = id
	}
	label = strings.Trim(label, `"`)

	if _, exists := participants[id]; exists {
		return true, fmt.Errorf("duplicate participant %q", id)
	}

	p := &Participant{
		ID:    id,
		Label: label,
		Index: len(sd.Participants),
	}
	sd.Participants = append(sd.Participants, p)
	participants[id] = p
	return true, nil
}

func (sd *SequenceDiagram) parseMessage(line string, participants map[string]*Participant) (bool, error) {
	match := messageRegex.FindStringSubmatch(line)
	if match == nil {
		return false, nil
	}

	fromID := match[2]
	if match[1] != "" {
		fromID = match[1]
	}

	arrow := match[3]

	toID := match[5]
	if match[4] != "" {
		toID = match[4]
	}

	label := strings.TrimSpace(match[6])

	from := sd.getParticipant(fromID, participants)
	to := sd.getParticipant(toID, participants)

	aType := DottedArrow
	if arrow == SolidArrowSyntax {
		aType = SolidArrow
	}

	msgNumber := 0
	if sd.Autonumber {
		msgNumber = len(sd.Messages) + 1
	}

	msg := &Message{
		From:      from,
		To:        to,
		Label:     label,
		ArrowType: aType,
		Number:    msgNumber,
	}
	sd.Messages = append(sd.Messages, msg)
	sd.Elements = append(sd.Elements, msg)
	return true, nil
}

func (sd *SequenceDiagram) getParticipant(id string, participants map[string]*Participant) *Participant {
	if p, exists := participants[id]; exists {
		return p
	}

	p := &Participant{
		ID:    id,
		Label: id,
		Index: len(sd.Participants),
	}
	sd.Participants = append(sd.Participants, p)
	participants[id] = p
	return p
}

func (sd *SequenceDiagram) parseNote(line string, participants map[string]*Participant) (bool, error) {
	match := noteRegex.FindStringSubmatch(line)
	if match == nil {
		return false, nil
	}

	posStr := strings.ToLower(match[1])
	actorsStr := strings.TrimSpace(match[2])
	text := strings.TrimSpace(match[3])

	var position NotePosition
	switch {
	case posStr == "over":
		position = NoteOver
	case strings.Contains(posStr, "left"):
		position = NoteLeftOf
	case strings.Contains(posStr, "right"):
		position = NoteRightOf
	default:
		return false, fmt.Errorf("unknown note position: %q", posStr)
	}

	// Parse actor(s) - comma separated for "over" with multiple actors
	actorIDs := strings.Split(actorsStr, ",")
	var actors []*Participant
	for _, id := range actorIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		actors = append(actors, sd.getParticipant(id, participants))
	}

	if len(actors) == 0 {
		return false, fmt.Errorf("note requires at least one actor")
	}

	if position != NoteOver && len(actors) > 1 {
		return false, fmt.Errorf("note %s only supports one actor", position)
	}

	note := &Note{
		Position: position,
		Actors:   actors,
		Text:     text,
	}
	sd.Elements = append(sd.Elements, note)
	return true, nil
}

func (sd *SequenceDiagram) parseBlock(lines []string, startIdx int, participants map[string]*Participant) (*Block, int, error) {
	if startIdx >= len(lines) {
		return nil, startIdx, fmt.Errorf("unexpected end of input")
	}

	match := blockStartRegex.FindStringSubmatch(lines[startIdx])
	if match == nil {
		return nil, startIdx, fmt.Errorf("expected block start")
	}

	blockType, err := parseBlockType(match[1])
	if err != nil {
		return nil, startIdx, err
	}

	block := &Block{
		Type:  blockType,
		Label: strings.TrimSpace(match[2]),
		Sections: []*BlockSection{
			{Label: "", Elements: []DiagramElement{}},
		},
	}

	currentSection := block.Sections[0]
	idx := startIdx + 1

	for idx < len(lines) {
		line := lines[idx]
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			idx++
			continue
		}

		if blockEndRegex.MatchString(trimmed) {
			return block, idx + 1, nil
		}

		if divMatch := blockDividerRegex.FindStringSubmatch(trimmed); divMatch != nil {
			divider := divMatch[1]
			if !isValidDivider(block.Type, divider) {
				return nil, idx, fmt.Errorf("invalid divider %q for block type %s", divider, block.Type)
			}
			currentSection = &BlockSection{
				Label:    strings.TrimSpace(divMatch[2]),
				Elements: []DiagramElement{},
			}
			block.Sections = append(block.Sections, currentSection)
			idx++
			continue
		}

		if blockStartRegex.MatchString(trimmed) {
			nestedBlock, nextIdx, err := sd.parseBlock(lines, idx, participants)
			if err != nil {
				return nil, idx, fmt.Errorf("nested block: %w", err)
			}
			currentSection.Elements = append(currentSection.Elements, nestedBlock)
			idx = nextIdx
			continue
		}

		if noteRegex.MatchString(trimmed) {
			elemsBefore := len(sd.Elements)
			if matched, err := sd.parseNote(trimmed, participants); err != nil {
				return nil, idx, err
			} else if matched {
				if len(sd.Elements) > elemsBefore {
					lastElem := sd.Elements[len(sd.Elements)-1]
					sd.Elements = sd.Elements[:len(sd.Elements)-1]
					currentSection.Elements = append(currentSection.Elements, lastElem)
				}
				idx++
				continue
			}
		}

		if messageRegex.MatchString(trimmed) {
			elemsBefore := len(sd.Elements)
			if matched, err := sd.parseMessage(trimmed, participants); err != nil {
				return nil, idx, err
			} else if matched {
				if len(sd.Elements) > elemsBefore {
					lastElem := sd.Elements[len(sd.Elements)-1]
					sd.Elements = sd.Elements[:len(sd.Elements)-1]
					currentSection.Elements = append(currentSection.Elements, lastElem)
				}
				idx++
				continue
			}
		}

		return nil, idx, fmt.Errorf("invalid syntax in block: %q", trimmed)
	}

	return nil, startIdx, fmt.Errorf("block starting at line %d has no 'end'", startIdx+1)
}
