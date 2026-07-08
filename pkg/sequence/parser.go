package sequence

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/AlexanderGrooff/mermaid-ascii/pkg/diagram"
)

const (
	SequenceDiagramKeyword = "sequenceDiagram"
	SolidArrowSyntax       = "->>"
	DottedArrowSyntax      = "-->>"
)

var (
	// participantRegex matches participant declarations: participant [ID] [as Label]
	participantRegex = regexp.MustCompile(`^\s*participant\s+(?:"([^"]+)"|(\S+))(?:\s+as\s+(.+))?$`)

	// messageRegex matches messages: [From][arrow][To]: [Label]. The arrow is one
	// of ->>, -->>, -> or -->. Unquoted participant names exclude the arrow
	// characters (- > <) so an unsupported arrow such as the bidirectional
	// "<<->>" cannot be silently absorbed into a name — it fails to match and is
	// reported as invalid syntax rather than rendered wrongly.
	messageRegex = regexp.MustCompile(`^\s*(?:"([^"]+)"|([^\s<>-]+))\s*(-->>|-->|->>|->)\s*(?:"([^"]+)"|([^\s<>-]+))\s*:\s*(.*)$`)

	// autonumberRegex matches the autonumber directive
	autonumberRegex = regexp.MustCompile(`^\s*autonumber\s*$`)

	// fragmentStartRegex matches the opening line of a control-flow fragment,
	// e.g. "loop every minute" or "opt is premium". Group 1 is the keyword,
	// group 2 is the (optional) label describing the condition.
	fragmentStartRegex = regexp.MustCompile(`^\s*(loop|opt)\b\s*(.*)$`)

	// fragmentEndRegex matches the "end" line that closes a fragment.
	fragmentEndRegex = regexp.MustCompile(`^\s*end\s*$`)
)

// SequenceDiagram represents a parsed sequence diagram.
type SequenceDiagram struct {
	Participants []*Participant
	// Messages is the flat list of every message arrow, in source order and
	// independent of any fragment nesting. Kept for callers that only care
	// about the messages themselves (and for backward compatibility).
	Messages []*Message
	// Events is the ordered body of the diagram used for rendering: each entry
	// is either a message or a fragment boundary. Walking Events reproduces the
	// original source order, including where loop/opt blocks open and close.
	Events     []Event
	Autonumber bool
}

// FragmentType identifies a control-flow fragment (a "framed" block of
// messages) such as a loop or an optional section.
type FragmentType int

const (
	FragmentLoop FragmentType = iota // loop ... end
	FragmentOpt                      // opt ... end
)

func (f FragmentType) String() string {
	switch f {
	case FragmentLoop:
		return "loop"
	case FragmentOpt:
		return "opt"
	default:
		return fmt.Sprintf("FragmentType(%d)", int(f))
	}
}

// Fragment describes the opening of a control-flow block: its kind and the
// optional condition text shown in the frame's label tab.
type Fragment struct {
	Type  FragmentType
	Label string
}

// EventKind tags each Event in the diagram body.
type EventKind int

const (
	EventMessage       EventKind = iota // a message arrow
	EventFragmentStart                  // the opening line of a loop/opt block
	EventFragmentEnd                    // the matching "end" line
)

// Event is one item in the diagram body. Exactly one payload field is set:
// Message when Kind is EventMessage, Fragment when Kind is EventFragmentStart.
// An EventFragmentEnd carries no payload; it just marks where a block closes.
type Event struct {
	Kind     EventKind
	Message  *Message
	Fragment *Fragment
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
	SolidArrow  ArrowType = iota // ->>  solid line with an arrowhead
	DottedArrow                  // -->> dotted line with an arrowhead
	SolidOpen                    // ->   solid line, no arrowhead
	DottedOpen                   // -->  dotted line, no arrowhead
)

// isDotted reports whether the arrow is drawn with a dotted (rather than solid)
// line.
func (a ArrowType) isDotted() bool {
	return a == DottedArrow || a == DottedOpen
}

// hasHead reports whether the arrow terminates in an arrowhead. The open forms
// (-> and -->) are drawn as a plain line touching the target lifeline.
func (a ArrowType) hasHead() bool {
	return a == SolidArrow || a == DottedArrow
}

func (a ArrowType) String() string {
	switch a {
	case SolidArrow:
		return "solid"
	case DottedArrow:
		return "dotted"
	case SolidOpen:
		return "solid-open"
	case DottedOpen:
		return "dotted-open"
	default:
		return fmt.Sprintf("ArrowType(%d)", int(a))
	}
}

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
	// openFragments counts how many loop/opt blocks are currently open so we can
	// reject an "end" with no matching opener and, at the very end, an opener
	// with no matching "end".
	openFragments := 0

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

		// Messages are checked before fragment keywords so a participant named
		// "loop"/"opt"/"end" (e.g. "loop->>B: hi") is still read as a message —
		// only bare openers like "loop retry" fall through to the checks below.
		if matched, err := sd.parseMessage(trimmed, participantMap); err != nil {
			return nil, fmt.Errorf("line %d: %w", i+2, err)
		} else if matched {
			continue
		}

		// A fragment opener ("loop"/"opt") starts a framed block.
		if match := fragmentStartRegex.FindStringSubmatch(trimmed); match != nil {
			fType := FragmentLoop
			if match[1] == "opt" {
				fType = FragmentOpt
			}
			sd.Events = append(sd.Events, Event{
				Kind:     EventFragmentStart,
				Fragment: &Fragment{Type: fType, Label: strings.TrimSpace(match[2])},
			})
			openFragments++
			continue
		}

		// "end" closes the most recently opened fragment.
		if fragmentEndRegex.MatchString(trimmed) {
			if openFragments == 0 {
				return nil, fmt.Errorf("line %d: %q without matching loop/opt", i+2, trimmed)
			}
			sd.Events = append(sd.Events, Event{Kind: EventFragmentEnd})
			openFragments--
			continue
		}

		return nil, fmt.Errorf("line %d: invalid syntax: %q", i+2, trimmed)
	}

	if openFragments > 0 {
		return nil, fmt.Errorf("unclosed loop/opt fragment: missing %d \"end\"", openFragments)
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

	var aType ArrowType
	switch arrow {
	case "->>":
		aType = SolidArrow
	case "-->>":
		aType = DottedArrow
	case "->":
		aType = SolidOpen
	case "-->":
		aType = DottedOpen
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
	sd.Events = append(sd.Events, Event{Kind: EventMessage, Message: msg})
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
