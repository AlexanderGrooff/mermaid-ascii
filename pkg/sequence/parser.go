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
	// e.g. "loop every minute", "opt is premium", "alt is valid". Group 1 is the
	// keyword, group 2 is the (optional) label describing the condition.
	fragmentStartRegex = regexp.MustCompile(`^\s*(loop|opt|alt)\b\s*(.*)$`)

	// fragmentElseRegex matches an "else" divider inside an alt block. Group 1 is
	// the (optional) condition label for the following section.
	fragmentElseRegex = regexp.MustCompile(`^\s*else\b\s*(.*)$`)

	// fragmentEndRegex matches the "end" line that closes a fragment.
	fragmentEndRegex = regexp.MustCompile(`^\s*end\s*$`)

	// noteRegex matches note annotations: "Note over A: text", "note left of A:
	// text", "Note over A,B: text" (case-insensitive keyword). Group 1 is the
	// placement, group 2 the participant list, group 3 the text.
	noteRegex = regexp.MustCompile(`^\s*[Nn]ote\s+(right of|left of|over)\s+([^:]+?)\s*:\s*(.*)$`)
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
	FragmentAlt                      // alt ... else ... end
)

func (f FragmentType) String() string {
	switch f {
	case FragmentLoop:
		return "loop"
	case FragmentOpt:
		return "opt"
	case FragmentAlt:
		return "alt"
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
	EventMessage         EventKind = iota // a message arrow
	EventFragmentStart                    // the opening line of a loop/opt/alt block
	EventFragmentDivider                  // an "else" section divider within an alt
	EventFragmentEnd                      // the matching "end" line
	EventNote                             // a note annotation
)

func (k EventKind) String() string {
	switch k {
	case EventMessage:
		return "message"
	case EventFragmentStart:
		return "fragment-start"
	case EventFragmentDivider:
		return "fragment-divider"
	case EventFragmentEnd:
		return "fragment-end"
	case EventNote:
		return "note"
	default:
		return fmt.Sprintf("EventKind(%d)", int(k))
	}
}

// Event is one item in the diagram body. Exactly one payload field is set:
// Message when Kind is EventMessage, Fragment when Kind is EventFragmentStart,
// Note when Kind is EventNote. An EventFragmentEnd carries no payload; it just
// marks where a block closes.
type Event struct {
	Kind     EventKind
	Message  *Message
	Fragment *Fragment
	Note     *Note
}

// NotePlacement describes where a note box sits relative to its participant(s).
type NotePlacement int

const (
	NoteOver    NotePlacement = iota // note over A  /  note over A,B
	NoteLeftOf                       // note left of A
	NoteRightOf                      // note right of A
)

// Note is an annotation box drawn over or beside participant lifelines.
type Note struct {
	Placement    NotePlacement
	Participants []*Participant // one participant, or two for "over A,B"
	Text         string
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
	// openFragments is a stack of the fragment types currently open, so we can
	// reject an "end"/"else" with no matching opener, validate that "else" only
	// appears inside an "alt", and detect an opener with no matching "end".
	var openFragments []FragmentType

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

		// Notes carry no arrow, so they never collide with messages; a
		// placement keyword is required, so a participant named "Note" (e.g.
		// "Note->>B: hi") still parses as a message further down.
		if m := noteRegex.FindStringSubmatch(trimmed); m != nil {
			placement := NoteOver
			switch m[1] {
			case "left of":
				placement = NoteLeftOf
			case "right of":
				placement = NoteRightOf
			}
			var parts []*Participant
			for _, id := range strings.Split(m[2], ",") {
				id = strings.Trim(strings.TrimSpace(id), `"`)
				if id != "" {
					parts = append(parts, sd.getParticipant(id, participantMap))
				}
			}
			if len(parts) == 0 {
				return nil, fmt.Errorf("line %d: note without a participant", i+2)
			}
			// Mermaid allows an optional wrap:/nowrap: prefix on note text;
			// wrapping is irrelevant for single-line ASCII, so just strip it.
			text := strings.TrimSpace(m[3])
			for _, pre := range []string{"nowrap:", "wrap:"} {
				if strings.HasPrefix(text, pre) {
					text = strings.TrimSpace(text[len(pre):])
					break
				}
			}
			sd.Events = append(sd.Events, Event{
				Kind: EventNote,
				Note: &Note{Placement: placement, Participants: parts, Text: text},
			})
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

		// A fragment opener ("loop"/"opt"/"alt") starts a framed block.
		if match := fragmentStartRegex.FindStringSubmatch(trimmed); match != nil {
			fType := map[string]FragmentType{"loop": FragmentLoop, "opt": FragmentOpt, "alt": FragmentAlt}[match[1]]
			sd.Events = append(sd.Events, Event{
				Kind:     EventFragmentStart,
				Fragment: &Fragment{Type: fType, Label: strings.TrimSpace(match[2])},
			})
			openFragments = append(openFragments, fType)
			continue
		}

		// "else" divides an alt block into sections.
		if match := fragmentElseRegex.FindStringSubmatch(trimmed); match != nil {
			if len(openFragments) == 0 || openFragments[len(openFragments)-1] != FragmentAlt {
				return nil, fmt.Errorf("line %d: %q outside an alt block", i+2, trimmed)
			}
			sd.Events = append(sd.Events, Event{
				Kind:     EventFragmentDivider,
				Fragment: &Fragment{Type: FragmentAlt, Label: strings.TrimSpace(match[1])},
			})
			continue
		}

		// "end" closes the most recently opened fragment.
		if fragmentEndRegex.MatchString(trimmed) {
			if len(openFragments) == 0 {
				return nil, fmt.Errorf("line %d: %q without matching loop/opt/alt", i+2, trimmed)
			}
			sd.Events = append(sd.Events, Event{Kind: EventFragmentEnd})
			openFragments = openFragments[:len(openFragments)-1]
			continue
		}

		return nil, fmt.Errorf("line %d: invalid syntax: %q", i+2, trimmed)
	}

	if len(openFragments) > 0 {
		return nil, fmt.Errorf("unclosed fragment: missing %d \"end\"", len(openFragments))
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
