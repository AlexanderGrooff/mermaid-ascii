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
	// participantRegex matches participant declarations: participant [rest] or
	// actor [rest]. mermaid renders an actor as a stick figure; in ASCII both
	// forms draw the same labelled box. The rest ([ID] [as Label]) is split by
	// parseParticipant, because IDs may contain spaces (mermaid's ID lexer
	// state allows them: "participant cron job as Cron").
	participantRegex = regexp.MustCompile(`(?i)^\s*(?:participant|actor)\s+(.+)$`)

	// participantAsRegex splits "ID as Label" at the FIRST " as ", matching
	// mermaid's lexer for whitespace-free IDs. For IDs containing spaces this
	// is a deliberate extension: mermaid's alias rule only fires for
	// whitespace-free IDs, so it would read the whole line as one long name.
	participantAsRegex = regexp.MustCompile(`(?i)^(.*?\S)\s+as\s+(.+)$`)

	// fragmentKeywordRegex guards message parsing: mermaid lexes fragment
	// keywords before actor names, so a line whose first word is one of them
	// is never a message — even when its label contains an arrow and a colon
	// ("else fall back -> retry: yes").
	fragmentKeywordRegex = regexp.MustCompile(`(?i)^(loop|opt|alt|par|critical|break|rect|else|and|option|end)\b`)

	// arrowTokens are mermaid's ten message arrows, longest first so that at
	// any position the longest token wins (e.g. "-->>" is never read as
	// "-->", "--)" never as "--" + ")").
	arrowTokens = []string{"<<-->>", "<<->>", "-->>", "--x", "--)", "-->", "->>", "-x", "-)", "->"}

	// autonumberRegex matches the autonumber directive
	autonumberRegex = regexp.MustCompile(`(?i)^\s*autonumber\s*$`)

	// fragmentStartRegex matches the opening line of a control-flow fragment,
	// e.g. "loop every minute", "opt is premium", "alt is valid". Group 1 is the
	// keyword, group 2 is the (optional) label describing the condition.
	fragmentStartRegex = regexp.MustCompile(`(?i)^\s*(loop|opt|alt|par|critical|break|rect)\b\s*(.*)$`)

	// fragmentDividerRegex matches a section divider inside a fragment: "else"
	// (alt), "and" (par), or "option" (critical). Group 1 is the keyword, group 2
	// the (optional) label for the following section.
	fragmentDividerRegex = regexp.MustCompile(`(?i)^\s*(else|and|option)\b\s*(.*)$`)

	// rectColorRegex strips a leading rgb()/rgba() colour argument from a rect's
	// label (ASCII can't render the fill; PR6 draws a plain frame).
	rectColorRegex = regexp.MustCompile(`(?i)^\s*rgba?\([^)]*\)\s*`)

	// fragmentEndRegex matches the "end" line that closes a fragment.
	fragmentEndRegex = regexp.MustCompile(`(?i)^\s*end\s*$`)

	// boxStartRegex matches a participant-group opener: `box [color] [title]`.
	// Like participant/actor (and mermaid itself), the keyword wins at line
	// start, so a participant named "box …" can only message via quoting.
	boxStartRegex = regexp.MustCompile(`(?i)^\s*box(?:\s+(.*))?$`)

	// boxColorFuncRegex matches a functional color prefix on a box line:
	// rgb(…), rgba(…), hsl(…), hsla(…) — the argument list may contain spaces.
	boxColorFuncRegex = regexp.MustCompile(`(?i)^(?:rgba?|hsla?)\s*\([^)]*\)`)

	// boxHexColorRegex matches a leading #hex color token.
	boxHexColorRegex = regexp.MustCompile(`^#[0-9a-fA-F]{3,8}\b`)

	// activationRegex matches the standalone activation keywords:
	// `activate A` / `deactivate A`. Group 1 is the keyword, group 2 the name.
	activationRegex = regexp.MustCompile(`(?i)^\s*(activate|deactivate)\s+(.+)$`)

	// brTagRegex matches <br> variants, which mermaid treats as line breaks;
	// single-line ASCII titles render them as spaces.
	brTagRegex = regexp.MustCompile(`(?i)<br\s*/?>`)

	// noteRegex matches note annotations: "Note over A: text", "note left of A:
	// text", "Note over A,B: text" (case-insensitive keyword). Group 1 is the
	// placement, group 2 the participant list, group 3 the text.
	noteRegex = regexp.MustCompile(`(?i)^\s*note\s+(right of|left of|over)\s+([^:]+?)\s*:\s*(.*)$`)
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
	// Boxes are participant groups declared with `box [color] [title] … end`;
	// each wraps a contiguous run of participants (they are declared inside
	// the block, so contiguity is inherent).
	Boxes []*Box
}

// Box is a participant group drawn as a frame around its participants'
// columns. mermaid fills it with a color; ASCII draws a plain frame, so the
// color is parsed (to keep it out of the title) and discarded.
type Box struct {
	Title       string
	First, Last int // participant Index range, inclusive
}

// FragmentType identifies a control-flow fragment (a "framed" block of
// messages) such as a loop or an optional section.
type FragmentType int

const (
	FragmentLoop     FragmentType = iota // loop ... end
	FragmentOpt                          // opt ... end
	FragmentAlt                          // alt ... else ... end
	FragmentPar                          // par ... and ... end
	FragmentCritical                     // critical ... option ... end
	FragmentBreak                        // break ... end
	FragmentRect                         // rect ... end
)

func (f FragmentType) String() string {
	switch f {
	case FragmentLoop:
		return "loop"
	case FragmentOpt:
		return "opt"
	case FragmentAlt:
		return "alt"
	case FragmentPar:
		return "par"
	case FragmentCritical:
		return "critical"
	case FragmentBreak:
		return "break"
	case FragmentRect:
		return "rect"
	default:
		return fmt.Sprintf("FragmentType(%d)", int(f))
	}
}

// fragmentKeywords maps an opener keyword to its fragment type.
var fragmentKeywords = map[string]FragmentType{
	"loop": FragmentLoop, "opt": FragmentOpt, "alt": FragmentAlt,
	"par": FragmentPar, "critical": FragmentCritical,
	"break": FragmentBreak, "rect": FragmentRect,
}

// dividerKeywords maps a section-divider keyword to the fragment type it must
// appear inside.
var dividerKeywords = map[string]FragmentType{
	"else": FragmentAlt, "and": FragmentPar, "option": FragmentCritical,
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
	EventActivate                         // a participant becomes active
	EventDeactivate                       // a participant stops being active
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
	case EventActivate:
		return "activate"
	case EventDeactivate:
		return "deactivate"
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
	// Participant is set for EventActivate / EventDeactivate: the lifeline
	// whose activation period starts or ends here.
	Participant *Participant
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
	// declared is true once a participant/actor statement names this
	// participant, as opposed to it being created implicitly by a message.
	declared bool
}

type Message struct {
	From      *Participant
	To        *Participant
	Label     string
	ArrowType ArrowType
	// CentralFrom / CentralTo mark central connections (mermaid's "()" on
	// either side of the arrow): a circle drawn where the message meets that
	// participant's lifeline.
	CentralFrom bool
	CentralTo   bool
	Number      int // Message number when autonumber is enabled (0 means no number)
}

type ArrowType int

const (
	SolidArrow          ArrowType = iota // ->>    solid line with an arrowhead
	DottedArrow                          // -->>   dotted line with an arrowhead
	SolidOpen                            // ->     solid line, no arrowhead
	DottedOpen                           // -->    dotted line, no arrowhead
	SolidCross                           // -x     solid line, cross head (lost/failed message)
	DottedCross                          // --x    dotted line, cross head
	SolidPoint                           // -)     solid line, open point head (async message)
	DottedPoint                          // --)    dotted line, open point head
	BidirectionalSolid                   // <<->>  solid line, arrowheads both ends
	BidirectionalDotted                  // <<-->> dotted line, arrowheads both ends
)

// isDotted reports whether the arrow is drawn with a dotted (rather than solid)
// line.
func (a ArrowType) isDotted() bool {
	switch a {
	case DottedArrow, DottedOpen, DottedCross, DottedPoint, BidirectionalDotted:
		return true
	}
	return false
}

// isBidirectional reports whether the arrow carries a head at the source end
// too (mermaid's <<->> and <<-->>).
func (a ArrowType) isBidirectional() bool {
	return a == BidirectionalSolid || a == BidirectionalDotted
}

// head returns the glyph drawn where the arrow meets the target lifeline, and
// false for the open forms (-> and -->), which are drawn as a plain line
// touching the lifeline. rightward selects the direction the head points.
func (a ArrowType) head(chars BoxChars, rightward bool) (rune, bool) {
	switch a {
	case SolidArrow, DottedArrow, BidirectionalSolid, BidirectionalDotted:
		if rightward {
			return chars.ArrowRight, true
		}
		return chars.ArrowLeft, true
	case SolidCross, DottedCross:
		return chars.CrossHead, true
	case SolidPoint, DottedPoint:
		if rightward {
			return chars.PointRight, true
		}
		return chars.PointLeft, true
	}
	return 0, false
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
	case SolidCross:
		return "solid-cross"
	case DottedCross:
		return "dotted-cross"
	case SolidPoint:
		return "solid-point"
	case DottedPoint:
		return "dotted-point"
	case BidirectionalSolid:
		return "bidirectional-solid"
	case BidirectionalDotted:
		return "bidirectional-dotted"
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
		return hasSequenceKeyword(trimmed)
	}
	return false
}

// hasSequenceKeyword reports whether a line is the sequenceDiagram declaration,
// case-insensitively (mermaid's sequence grammar is case-insensitive). The
// keyword must stand as a whole token — followed by whitespace or end of line —
// so a node id like "sequenceDiagramFoo" in a flowchart isn't misrouted here.
func hasSequenceKeyword(line string) bool {
	lower := strings.ToLower(strings.TrimSpace(line))
	kw := strings.ToLower(SequenceDiagramKeyword)
	if !strings.HasPrefix(lower, kw) {
		return false
	}
	rest := lower[len(kw):]
	return rest == "" || rest[0] == ' ' || rest[0] == '\t'
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

	if !hasSequenceKeyword(strings.TrimSpace(lines[0])) {
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
	// openBox is the box block currently being declared, if any. mermaid's
	// grammar allows only participant declarations inside a box, and boxes
	// cannot nest.
	var openBox *Box
	// active counts each participant's open activation periods, so deactivating
	// an inactive participant is an error and stacked activations balance.
	active := map[*Participant]int{}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Inside a box block only participant/actor declarations (and the
		// closing "end") are valid, so handle that mode first.
		if openBox != nil {
			if fragmentEndRegex.MatchString(trimmed) {
				openBox = nil
				continue
			}
			if boxStartRegex.MatchString(trimmed) {
				return nil, fmt.Errorf("line %d: boxes cannot nest", i+2)
			}
			p, matched, err := sd.parseParticipant(trimmed, participantMap)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", i+2, err)
			}
			if !matched {
				return nil, fmt.Errorf("line %d: only participant declarations are allowed inside a box: %q", i+2, trimmed)
			}
			if openBox.First == -1 {
				openBox.First = p.Index
			}
			openBox.Last = p.Index
			continue
		}

		// Check for autonumber directive
		if autonumberRegex.MatchString(trimmed) {
			sd.Autonumber = true
			continue
		}

		// A box opener starts a participant group; its optional color is
		// parsed away (no ASCII meaning) so it never leaks into the title.
		if m := boxStartRegex.FindStringSubmatch(trimmed); m != nil {
			openBox = &Box{Title: parseBoxTitle(m[1]), First: -1}
			sd.Boxes = append(sd.Boxes, openBox)
			continue
		}

		// Notes carry no arrow, so they never collide with messages; a
		// placement keyword is required, so a participant named "Note" (e.g.
		// "Note->>B: hi") still parses as a message further down.
		if m := noteRegex.FindStringSubmatch(trimmed); m != nil {
			placement := NoteOver
			switch strings.ToLower(m[1]) { // keyword may be any case
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
				if strings.HasPrefix(strings.ToLower(text), pre) {
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

		if _, matched, err := sd.parseParticipant(trimmed, participantMap); err != nil {
			return nil, fmt.Errorf("line %d: %w", i+2, err)
		} else if matched {
			continue
		}

		// Messages are checked before fragment keywords so a participant named
		// "loop"/"opt"/"end" (e.g. "loop->>B: hi") is still read as a message —
		// only bare openers like "loop retry" fall through to the checks below.
		if matched, err := sd.parseMessage(trimmed, participantMap, active); err != nil {
			return nil, fmt.Errorf("line %d: %w", i+2, err)
		} else if matched {
			continue
		}

		// Standalone activation keywords. Checked after messages so a
		// participant named "activate …" can still send one.
		if m := activationRegex.FindStringSubmatch(trimmed); m != nil {
			name, nameOK := parseName(m[2])
			if !nameOK {
				return nil, fmt.Errorf("line %d: invalid participant name %q", i+2, m[2])
			}
			p := sd.getParticipant(name, participantMap)
			if strings.EqualFold(m[1], "activate") {
				active[p]++
				sd.Events = append(sd.Events, Event{Kind: EventActivate, Participant: p})
			} else {
				if active[p] == 0 {
					return nil, fmt.Errorf("line %d: trying to deactivate an inactive participant %q", i+2, p.ID)
				}
				active[p]--
				sd.Events = append(sd.Events, Event{Kind: EventDeactivate, Participant: p})
			}
			continue
		}

		// A fragment opener (loop/opt/alt/par/critical/break/rect) starts a block.
		if match := fragmentStartRegex.FindStringSubmatch(trimmed); match != nil {
			fType := fragmentKeywords[strings.ToLower(match[1])]
			label := strings.TrimSpace(match[2])
			// rect's argument is a fill colour we can't render in ASCII; drop it
			// so the frame is drawn plain (colour support is a follow-up).
			if fType == FragmentRect {
				label = strings.TrimSpace(rectColorRegex.ReplaceAllString(label, ""))
			}
			sd.Events = append(sd.Events, Event{
				Kind:     EventFragmentStart,
				Fragment: &Fragment{Type: fType, Label: label},
			})
			openFragments = append(openFragments, fType)
			continue
		}

		// A section divider: "else" (alt), "and" (par), "option" (critical). It
		// must sit directly inside the matching fragment type.
		if match := fragmentDividerRegex.FindStringSubmatch(trimmed); match != nil {
			want := dividerKeywords[strings.ToLower(match[1])]
			if len(openFragments) == 0 || openFragments[len(openFragments)-1] != want {
				return nil, fmt.Errorf("line %d: %q outside a matching %s block", i+2, trimmed, want)
			}
			sd.Events = append(sd.Events, Event{
				Kind:     EventFragmentDivider,
				Fragment: &Fragment{Type: want, Label: strings.TrimSpace(match[2])},
			})
			continue
		}

		// "end" closes the most recently opened fragment.
		if fragmentEndRegex.MatchString(trimmed) {
			if len(openFragments) == 0 {
				return nil, fmt.Errorf("line %d: %q without a matching fragment opener", i+2, trimmed)
			}
			sd.Events = append(sd.Events, Event{Kind: EventFragmentEnd})
			openFragments = openFragments[:len(openFragments)-1]
			continue
		}

		return nil, fmt.Errorf("line %d: invalid syntax: %q", i+2, trimmed)
	}

	if openBox != nil {
		return nil, fmt.Errorf("unclosed box: missing \"end\"")
	}
	if len(openFragments) > 0 {
		return nil, fmt.Errorf("unclosed fragment: missing %d \"end\"", len(openFragments))
	}

	if len(sd.Participants) == 0 {
		return nil, fmt.Errorf("no participants found")
	}

	// A box declared with no participants has nothing to frame; drop it.
	kept := sd.Boxes[:0]
	for _, b := range sd.Boxes {
		if b.First >= 0 {
			kept = append(kept, b)
		}
	}
	sd.Boxes = kept

	return sd, nil
}

// parseParticipant handles a participant/actor declaration. It reports whether
// the line was one, and returns the participant it created or claimed so a
// caller (the box block) can record which lifeline was declared.
func (sd *SequenceDiagram) parseParticipant(line string, participants map[string]*Participant) (*Participant, bool, error) {
	match := participantRegex.FindStringSubmatch(line)
	if match == nil {
		return nil, false, nil
	}

	rest := strings.TrimSpace(match[1])
	id := rest
	label := ""
	if asMatch := participantAsRegex.FindStringSubmatch(rest); asMatch != nil {
		id, label = asMatch[1], asMatch[2]
	}
	id, idOK := parseName(id)
	if !idOK {
		return nil, true, fmt.Errorf("invalid participant name %q", id)
	}
	if label == "" {
		label = id
	}
	label = strings.Trim(label, `"`)

	// A participant already created implicitly by an earlier message may be
	// declared afterwards to give it a label or put it in a box, as mermaid
	// allows (addActor updates an existing actor). Only declaring the same
	// participant twice is an error.
	if existing, exists := participants[id]; exists {
		if existing.declared {
			return nil, true, fmt.Errorf("duplicate participant %q", id)
		}
		existing.declared = true
		existing.Label = label
		return existing, true, nil
	}

	p := &Participant{
		ID:       id,
		Label:    label,
		Index:    len(sd.Participants),
		declared: true,
	}
	sd.Participants = append(sd.Participants, p)
	participants[id] = p
	return p, true, nil
}

// parseBoxTitle extracts the display title from a box opener's argument:
// an optional leading color (functional rgb()/hsl(), #hex, or a CSS color
// name — mermaid fills the box with it, ASCII has no use for it) followed by
// the title text. <br> tags become spaces, as elsewhere in single-line ASCII.
func parseBoxTitle(rest string) string {
	rest = strings.TrimSpace(brTagRegex.ReplaceAllString(rest, " "))
	if loc := boxColorFuncRegex.FindStringIndex(rest); loc != nil {
		return strings.TrimSpace(rest[loc[1]:])
	}
	if loc := boxHexColorRegex.FindStringIndex(rest); loc != nil {
		return strings.TrimSpace(rest[loc[1]:])
	}
	tok := rest
	if idx := strings.IndexAny(rest, " \t"); idx >= 0 {
		tok = rest[:idx]
	}
	if cssColorNames[strings.ToLower(tok)] {
		return strings.TrimSpace(rest[len(tok):])
	}
	return rest
}

// cssColorNames is the set of CSS color keywords (plus "transparent") that
// mermaid accepts as a box fill. Needed to split "box Green New serwis"
// (color Green, title "New serwis") from "box Facade" (no color, title
// "Facade") the way mermaid does.
var cssColorNames = func() map[string]bool {
	const names = "aliceblue antiquewhite aqua aquamarine azure beige bisque black blanchedalmond blue " +
		"blueviolet brown burlywood cadetblue chartreuse chocolate coral cornflowerblue cornsilk crimson " +
		"cyan darkblue darkcyan darkgoldenrod darkgray darkgreen darkgrey darkkhaki darkmagenta " +
		"darkolivegreen darkorange darkorchid darkred darksalmon darkseagreen darkslateblue darkslategray " +
		"darkslategrey darkturquoise darkviolet deeppink deepskyblue dimgray dimgrey dodgerblue firebrick " +
		"floralwhite forestgreen fuchsia gainsboro ghostwhite gold goldenrod gray green greenyellow grey " +
		"honeydew hotpink indianred indigo ivory khaki lavender lavenderblush lawngreen lemonchiffon " +
		"lightblue lightcoral lightcyan lightgoldenrodyellow lightgray lightgreen lightgrey lightpink " +
		"lightsalmon lightseagreen lightskyblue lightslategray lightslategrey lightsteelblue lightyellow " +
		"lime limegreen linen magenta maroon mediumaquamarine mediumblue mediumorchid mediumpurple " +
		"mediumseagreen mediumslateblue mediumspringgreen mediumturquoise mediumvioletred midnightblue " +
		"mintcream mistyrose moccasin navajowhite navy oldlace olive olivedrab orange orangered orchid " +
		"palegoldenrod palegreen paleturquoise palevioletred papayawhip peachpuff peru pink plum " +
		"powderblue purple rebeccapurple red rosybrown royalblue saddlebrown salmon sandybrown seagreen " +
		"seashell sienna silver skyblue slateblue slategray slategrey snow springgreen steelblue tan teal " +
		"thistle tomato turquoise violet wheat white whitesmoke yellow yellowgreen transparent"
	set := map[string]bool{}
	for _, n := range strings.Fields(names) {
		set[n] = true
	}
	return set
}()

// findArrow returns the index and token of the first arrow in line, skipping
// quoted spans. This mirrors mermaid's lexer, where a name may contain '-'
// (Alice-in-Wonderland) or spaces (cron job) because only a '-' or '<' that
// starts a real arrow token can end it.
func findArrow(line string) (int, string) {
	inQuotes := false
	for i := 0; i < len(line); i++ {
		switch {
		case line[i] == '"':
			inQuotes = !inQuotes
		case !inQuotes && (line[i] == '-' || line[i] == '<'):
			for _, tok := range arrowTokens {
				if strings.HasPrefix(line[i:], tok) {
					return i, tok
				}
			}
		}
	}
	return -1, ""
}

// validName reports whether an unquoted participant name is acceptable.
// Mermaid's ACTOR token excludes these characters; anything else (spaces,
// dashes not forming an arrow, '=', '.', ...) is a legal name.
func validName(name string) bool {
	return name != "" && !strings.ContainsAny(name, `"<>:,;(`)
}

// parseName extracts a participant name: either a quoted string or a bare
// name validated by validName.
func parseName(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if len(raw) >= 2 && strings.HasPrefix(raw, `"`) && strings.HasSuffix(raw, `"`) {
		inner := raw[1 : len(raw)-1]
		return inner, inner != "" && !strings.Contains(inner, `"`)
	}
	return raw, validName(raw)
}

// splitMessage breaks a line into [From][()][arrow][()][To]: [Label], where
// "()" marks a central connection on that side. ok is false when the line is
// not a message, so parsing can fall through to the other statement forms.
func splitMessage(line string) (m messageParts, ok bool) {
	idx, arrow := findArrow(line)
	if idx < 0 {
		return
	}
	m.arrow = arrow

	left := strings.TrimSpace(line[:idx])
	// A line opening with a fragment keyword is a fragment statement whose
	// label happens to contain an arrow, never a message. Quoted names are
	// exempt: `"loop svc" ->> B` is an explicit participant reference.
	if !strings.HasPrefix(left, `"`) && fragmentKeywordRegex.MatchString(left) {
		return
	}
	if strings.HasSuffix(left, "()") {
		m.centralFrom = true
		left = strings.TrimSpace(strings.TrimSuffix(left, "()"))
	}
	fromID, fromOK := parseName(left)
	m.fromID = fromID

	// A '+' or '-' after the arrow is the activation shorthand: '+' activates
	// the receiver, '-' deactivates the sender. mermaid's lexer skips
	// whitespace and no name may begin with a sign, so the sign is recognised
	// either side of a space.
	rest := strings.TrimLeft(line[idx+len(arrow):], " \t")
	if rest != "" && (rest[0] == '+' || rest[0] == '-') {
		m.activateTo = rest[0] == '+'
		m.deactivateFrom = rest[0] == '-'
		rest = rest[1:]
	}
	rest = strings.TrimSpace(rest)
	if strings.HasPrefix(rest, "()") {
		m.centralTo = true
		rest = strings.TrimSpace(strings.TrimPrefix(rest, "()"))
	}
	// The label starts at the first ':' after the (possibly quoted) to-name.
	start := 0
	if strings.HasPrefix(rest, `"`) {
		if end := strings.Index(rest[1:], `"`); end >= 0 {
			start = end + 2
		}
	}
	colon := strings.Index(rest[start:], ":")
	if colon < 0 {
		return
	}
	colon += start
	toID, toOK := parseName(rest[:colon])
	m.toID = toID
	m.label = strings.TrimSpace(rest[colon+1:])

	ok = fromOK && toOK
	return
}

// messageParts is the decomposition of a message line: endpoints, arrow, label,
// central-connection markers and the activation shorthand suffix.
type messageParts struct {
	fromID, arrow, toID, label string
	centralFrom, centralTo     bool
	activateTo                 bool // "+" after the arrow: activate the receiver
	deactivateFrom             bool // "-" after the arrow: deactivate the sender
}

func (sd *SequenceDiagram) parseMessage(line string, participants map[string]*Participant, active map[*Participant]int) (bool, error) {
	parts, ok := splitMessage(line)
	if !ok {
		return false, nil
	}

	from := sd.getParticipant(parts.fromID, participants)
	to := sd.getParticipant(parts.toID, participants)

	var aType ArrowType
	switch parts.arrow {
	case "->>":
		aType = SolidArrow
	case "-->>":
		aType = DottedArrow
	case "->":
		aType = SolidOpen
	case "-->":
		aType = DottedOpen
	case "-x":
		aType = SolidCross
	case "--x":
		aType = DottedCross
	case "-)":
		aType = SolidPoint
	case "--)":
		aType = DottedPoint
	case "<<->>":
		aType = BidirectionalSolid
	case "<<-->>":
		aType = BidirectionalDotted
	}

	msgNumber := 0
	if sd.Autonumber {
		msgNumber = len(sd.Messages) + 1
	}

	msg := &Message{
		From:        from,
		To:          to,
		Label:       parts.label,
		ArrowType:   aType,
		CentralFrom: parts.centralFrom,
		CentralTo:   parts.centralTo,
		Number:      msgNumber,
	}
	sd.Messages = append(sd.Messages, msg)
	sd.Events = append(sd.Events, Event{Kind: EventMessage, Message: msg})

	// Activation shorthand: the period starts (or ends) after the message that
	// carries the suffix, so the arrow itself sits on the boundary.
	if parts.activateTo {
		active[to]++
		sd.Events = append(sd.Events, Event{Kind: EventActivate, Participant: to})
	}
	if parts.deactivateFrom {
		if active[from] == 0 {
			return true, fmt.Errorf("trying to deactivate an inactive participant %q", from.ID)
		}
		active[from]--
		sd.Events = append(sd.Events, Event{Kind: EventDeactivate, Participant: from})
	}
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
