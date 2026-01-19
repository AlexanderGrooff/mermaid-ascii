package sequence

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/AlexanderGrooff/mermaid-ascii/internal/diagram"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		wantParticipants int
		wantMessages     int
		wantErr          string
	}{
		{"empty input", "", 0, 0, "empty input"},
		{"missing sequenceDiagram keyword", "A->>B: Hello", 0, 0, "expected \"sequenceDiagram\" keyword"},
		{"only comments", "sequenceDiagram\n%% This is a comment\n%% Another comment", 0, 0, "no participants found"},
		{"no participants", "sequenceDiagram", 0, 0, "no participants found"},
		{"duplicate participant ID", "sequenceDiagram\nparticipant Alice\nparticipant Alice\nAlice->>Bob: Hi", 0, 0, "duplicate participant"},
		{"minimal diagram", "sequenceDiagram\nA->>B: Hello", 2, 1, ""},
		{"explicit participants", "sequenceDiagram\nparticipant Alice\nparticipant Bob\nAlice->>Bob: Hi", 2, 1, ""},
		{"dotted arrow", "sequenceDiagram\nA-->>B: Response", 2, 1, ""},
		{"self message", "sequenceDiagram\nA->>A: Self", 1, 1, ""},
		{"multiple messages", "sequenceDiagram\nA->>B: 1\nB->>C: 2\nC-->>A: 3", 3, 3, ""},
		{"with comments", "sequenceDiagram\n%% Comment\nA->>B: Hi %% inline comment", 2, 1, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sd, err := Parse(tt.input)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("Expected error containing %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if len(sd.Participants) != tt.wantParticipants {
				t.Errorf("Expected %d participants, got %d", tt.wantParticipants, len(sd.Participants))
			}
			if len(sd.Messages) != tt.wantMessages {
				t.Errorf("Expected %d messages, got %d", tt.wantMessages, len(sd.Messages))
			}
		})
	}
}

func TestIsSequenceDiagram(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"sequenceDiagram\nA->>B: Hello", true},
		{"graph LR\nA-->B", false},
		{"graph TD\nA-->B", false},
		{"", false},
		{"%% Just a comment", false},
	}

	for _, tt := range tests {
		if got := IsSequenceDiagram(tt.input); got != tt.want {
			t.Errorf("IsSequenceDiagram(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParticipantAlias(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantID    string
		wantLabel string
	}{
		{"simple alias", "sequenceDiagram\nparticipant A as Alice\nA->>A: Hello", "A", "Alice"},
		{"no alias defaults to id", "sequenceDiagram\nparticipant Alice\nAlice->>Alice: Hi", "Alice", "Alice"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(d.Participants) == 0 {
				t.Fatal("expected at least one participant")
			}
			p := d.Participants[0]
			if p.ID != tt.wantID {
				t.Errorf("ID = %q, want %q", p.ID, tt.wantID)
			}
			if p.Label != tt.wantLabel {
				t.Errorf("Label = %q, want %q", p.Label, tt.wantLabel)
			}
			config := diagram.DefaultConfig()
			output, err := Render(d, config)
			if err != nil {
				t.Fatalf("render error: %v", err)
			}
			if !strings.Contains(output, tt.wantLabel) {
				t.Errorf("output should contain label %q", tt.wantLabel)
			}
		})
	}
}

func TestMessageRegex(t *testing.T) {
	tests := []struct {
		input     string
		wantFrom  string
		wantArrow string
		wantTo    string
		wantLabel string
		wantMatch bool
	}{
		{"A->>B: Hello", "A", "->>", "B", "Hello", true},
		{"A-->>B: Response", "A", "-->>", "B", "Response", true},
		{`"My Service"->>B: Test`, "My Service", "->>", "B", "Test", true},
		{"A->>B: ", "A", "->>", "B", "", true},
		{"A->B: Test", "", "", "", "", false},
		{"A->>B", "", "", "", "", false},
	}

	for _, tt := range tests {
		match := messageRegex.FindStringSubmatch(tt.input)
		if !tt.wantMatch {
			if match != nil {
				t.Errorf("messageRegex should not match %q", tt.input)
			}
			continue
		}
		if match == nil {
			t.Fatalf("messageRegex failed to match: %q", tt.input)
		}
		gotFrom := match[2]
		if match[1] != "" {
			gotFrom = match[1]
		}
		gotArrow := match[3]
		gotTo := match[5]
		if match[4] != "" {
			gotTo = match[4]
		}
		gotLabel := match[6]

		if gotFrom != tt.wantFrom || gotArrow != tt.wantArrow || gotTo != tt.wantTo || gotLabel != tt.wantLabel {
			t.Errorf("messageRegex(%q) = (%q, %q, %q, %q), want (%q, %q, %q, %q)",
				tt.input, gotFrom, gotArrow, gotTo, gotLabel, tt.wantFrom, tt.wantArrow, tt.wantTo, tt.wantLabel)
		}
	}
}

func TestParticipantRegex(t *testing.T) {
	tests := []struct {
		input     string
		wantID    string
		wantAlias string
	}{
		{"participant Alice", "Alice", ""},
		{"participant Alice as A", "Alice", "A"},
		{`participant "My Service"`, "My Service", ""},
		{`participant "My Service" as Service`, "My Service", "Service"},
	}

	for _, tt := range tests {
		match := participantRegex.FindStringSubmatch(tt.input)
		if match == nil {
			t.Fatalf("participantRegex failed to match: %q", tt.input)
		}
		gotID := match[2]
		if match[1] != "" {
			gotID = match[1]
		}
		gotAlias := match[3]

		if gotID != tt.wantID || gotAlias != tt.wantAlias {
			t.Errorf("participantRegex(%q) = (%q, %q), want (%q, %q)",
				tt.input, gotID, gotAlias, tt.wantID, tt.wantAlias)
		}
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"A->>B: Hello", []string{"A->>B: Hello"}},
		{"line1\nline2\nline3", []string{"line1", "line2", "line3"}},
		{"line1\\nline2\\nline3", []string{"line1", "line2", "line3"}},
		{"", []string{""}},
	}

	for _, tt := range tests {
		result := diagram.SplitLines(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("SplitLines(%q) len = %d, want %d", tt.input, len(result), len(tt.expected))
		}
	}
}

func TestRemoveComments(t *testing.T) {
	tests := []struct {
		input    []string
		expected []string
	}{
		{[]string{"A->>B: Hello", "B-->>A: Hi"}, []string{"A->>B: Hello", "B-->>A: Hi"}},
		{[]string{"%% This is a comment", "A->>B: Hello"}, []string{"A->>B: Hello"}},
		{[]string{"A->>B: Hello %% inline comment", "B-->>A: Hi"}, []string{"A->>B: Hello", "B-->>A: Hi"}},
	}

	for _, tt := range tests {
		result := diagram.RemoveComments(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("RemoveComments() len = %d, want %d", len(result), len(tt.expected))
		}
	}
}

func TestArrowTypeString(t *testing.T) {
	if SolidArrow.String() != "solid" {
		t.Errorf("SolidArrow.String() = %q, want \"solid\"", SolidArrow.String())
	}
	if DottedArrow.String() != "dotted" {
		t.Errorf("DottedArrow.String() = %q, want \"dotted\"", DottedArrow.String())
	}
}

func TestParseNoteQuotedActorSameAsMessage(t *testing.T) {
	input := `sequenceDiagram
		"My Service"->>B: Hello
		Note over "My Service": This is a note`

	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sd.Participants) != 2 {
		t.Fatalf("expected 2 participants, got %d: %v", len(sd.Participants), sd.Participants)
	}

	if sd.Participants[0].ID != "My Service" {
		t.Errorf("expected first participant ID to be 'My Service', got %q", sd.Participants[0].ID)
	}

	if len(sd.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(sd.Elements))
	}

	msg, ok := sd.Elements[0].(*Message)
	if !ok {
		t.Fatalf("expected Message, got %T", sd.Elements[0])
	}

	note, ok := sd.Elements[1].(*Note)
	if !ok {
		t.Fatalf("expected Note, got %T", sd.Elements[1])
	}

	if msg.From != note.Actors[0] {
		t.Errorf("message From participant (%p, ID=%q) should be same as note actor (%p, ID=%q)",
			msg.From, msg.From.ID, note.Actors[0], note.Actors[0].ID)
	}
}

func TestParseNoteOverSingleActor(t *testing.T) {
	input := `sequenceDiagram
		participant A
		Note over A: This is a note`

	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sd.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(sd.Elements))
	}

	note, ok := sd.Elements[0].(*Note)
	if !ok {
		t.Fatalf("expected Note, got %T", sd.Elements[0])
	}

	if note.Position != NoteOver {
		t.Errorf("expected NoteOver, got %v", note.Position)
	}
	if len(note.Actors) != 1 || note.Actors[0].ID != "A" {
		t.Errorf("expected 1 actor with ID 'A', got %v", note.Actors)
	}
	if note.Text != "This is a note" {
		t.Errorf("expected text 'This is a note', got %q", note.Text)
	}
}

func TestParseNoteOverMultipleActors(t *testing.T) {
	input := `sequenceDiagram
		participant A
		participant B
		Note over A,B: Spanning note`

	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sd.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(sd.Elements))
	}

	note, ok := sd.Elements[0].(*Note)
	if !ok {
		t.Fatalf("expected Note, got %T", sd.Elements[0])
	}

	if note.Position != NoteOver {
		t.Errorf("expected NoteOver, got %v", note.Position)
	}
	if len(note.Actors) != 2 {
		t.Fatalf("expected 2 actors, got %d", len(note.Actors))
	}
	if note.Actors[0].ID != "A" || note.Actors[1].ID != "B" {
		t.Errorf("expected actors A and B, got %v and %v", note.Actors[0].ID, note.Actors[1].ID)
	}
	if note.Text != "Spanning note" {
		t.Errorf("expected text 'Spanning note', got %q", note.Text)
	}
}

func TestRenderNoteOver(t *testing.T) {
	input := `sequenceDiagram
		participant A
		participant B
		Note over A: Test note`

	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	output, err := Render(sd, nil)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	if !strings.Contains(output, "Test note") {
		t.Errorf("output should contain note text:\n%s", output)
	}
}

func TestRenderNoteOverLongText(t *testing.T) {
	input := `sequenceDiagram
		participant A
		participant B
		Note over A: This is a very long note text that should expand the box`

	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	output, err := Render(sd, nil)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	longText := "This is a very long note text that should expand the box"
	if !strings.Contains(output, longText) {
		t.Errorf("output should contain full note text:\n%s", output)
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "TopLeft") || strings.Contains(line, "TopRight") {
			continue
		}
		for i, r := range line {
			if r == '│' || r == '|' {
				if i > 0 && i < len(line)-1 {
					prev := rune(line[i-1])
					next := rune(line[i+1])
					if (prev >= 'a' && prev <= 'z') || (prev >= 'A' && prev <= 'Z') {
						t.Errorf("border character at position %d may be overwriting text: %s", i, line)
					}
					if (next >= 'a' && next <= 'z') || (next >= 'A' && next <= 'Z') {
						if next != 'T' {
							t.Errorf("border character at position %d may be adjacent to truncated text: %s", i, line)
						}
					}
				}
			}
		}
	}
}

func TestParseNoteLeftRight(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantPosition NotePosition
		wantActorID  string
		wantText     string
	}{
		{
			name: "note left of",
			input: `sequenceDiagram
				participant A
				Note left of A: Left note`,
			wantPosition: NoteLeftOf,
			wantActorID:  "A",
			wantText:     "Left note",
		},
		{
			name: "note right of",
			input: `sequenceDiagram
				participant B
				Note right of B: Right note`,
			wantPosition: NoteRightOf,
			wantActorID:  "B",
			wantText:     "Right note",
		},
		{
			name: "note left of case insensitive",
			input: `sequenceDiagram
				participant C
				NOTE LEFT OF C: Case test`,
			wantPosition: NoteLeftOf,
			wantActorID:  "C",
			wantText:     "Case test",
		},
		{
			name: "note right of case insensitive",
			input: `sequenceDiagram
				participant D
				note RIGHT OF D: Mixed case`,
			wantPosition: NoteRightOf,
			wantActorID:  "D",
			wantText:     "Mixed case",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sd, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(sd.Elements) != 1 {
				t.Fatalf("expected 1 element, got %d", len(sd.Elements))
			}

			note, ok := sd.Elements[0].(*Note)
			if !ok {
				t.Fatalf("expected Note, got %T", sd.Elements[0])
			}

			if note.Position != tt.wantPosition {
				t.Errorf("expected position %v, got %v", tt.wantPosition, note.Position)
			}
			if len(note.Actors) != 1 || note.Actors[0].ID != tt.wantActorID {
				t.Errorf("expected actor %q, got %v", tt.wantActorID, note.Actors)
			}
			if note.Text != tt.wantText {
				t.Errorf("expected text %q, got %q", tt.wantText, note.Text)
			}
		})
	}
}

func TestRenderNoteRightOf(t *testing.T) {
	input := `sequenceDiagram
		participant A
		participant B
		Note right of B: Right note`

	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	output, err := Render(sd, nil)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	if !strings.Contains(output, "Right note") {
		t.Errorf("output should contain note text:\n%s", output)
	}
}

func TestRenderNoteRightOfEdgeBoundary(t *testing.T) {
	input := `sequenceDiagram
		participant A
		Note right of A: Hi`

	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	output, err := Render(sd, nil)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	if !strings.Contains(output, "Hi") {
		t.Errorf("output should contain note text:\n%s", output)
	}
}

func TestRenderNoteLeftOf(t *testing.T) {
	input := `sequenceDiagram
		participant A
		participant B
		Note left of A: Left note`

	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	output, err := Render(sd, nil)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	if !strings.Contains(output, "Left note") {
		t.Errorf("output should contain note text:\n%s", output)
	}
}

func TestParseBlockLoop(t *testing.T) {
	input := `sequenceDiagram
		participant A
		participant B
		loop Every minute
			A->>B: Ping
			B-->>A: Pong
		end`

	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sd.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(sd.Elements))
	}

	block, ok := sd.Elements[0].(*Block)
	if !ok {
		t.Fatalf("expected Block, got %T", sd.Elements[0])
	}

	if block.Type != BlockLoop {
		t.Errorf("expected BlockLoop, got %v", block.Type)
	}
	if block.Label != "Every minute" {
		t.Errorf("expected label 'Every minute', got %q", block.Label)
	}
	if len(block.Sections) != 1 {
		t.Errorf("expected 1 section, got %d", len(block.Sections))
	}
	if len(block.Sections[0].Elements) != 2 {
		t.Errorf("expected 2 elements in section, got %d", len(block.Sections[0].Elements))
	}
}

func TestParseBlockAltElse(t *testing.T) {
	input := `sequenceDiagram
		participant A
		participant B
		alt Success
			A->>B: 200 OK
		else Failure
			A->>B: 500 Error
		end`

	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	block, ok := sd.Elements[0].(*Block)
	if !ok {
		t.Fatalf("expected Block, got %T", sd.Elements[0])
	}

	if block.Type != BlockAlt {
		t.Errorf("expected BlockAlt, got %v", block.Type)
	}
	if len(block.Sections) != 2 {
		t.Errorf("expected 2 sections, got %d", len(block.Sections))
	}
	if block.Sections[1].Label != "Failure" {
		t.Errorf("expected section label 'Failure', got %q", block.Sections[1].Label)
	}
}

func TestParseBlockParAnd(t *testing.T) {
	input := `sequenceDiagram
		participant A
		participant B
		participant C
		par Task 1
			A->>B: Do X
		and Task 2
			A->>C: Do Y
		end`

	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	block, ok := sd.Elements[0].(*Block)
	if !ok {
		t.Fatalf("expected Block, got %T", sd.Elements[0])
	}

	if block.Type != BlockPar {
		t.Errorf("expected BlockPar, got %v", block.Type)
	}
	if len(block.Sections) != 2 {
		t.Errorf("expected 2 sections, got %d", len(block.Sections))
	}
}

func TestParseBlockNested(t *testing.T) {
	input := `sequenceDiagram
		participant A
		participant B
		loop Outer
			alt Check
				A->>B: Request
			else Skip
				A->>B: Skip
			end
		end`

	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outerBlock, ok := sd.Elements[0].(*Block)
	if !ok {
		t.Fatalf("expected Block, got %T", sd.Elements[0])
	}

	if outerBlock.Type != BlockLoop {
		t.Errorf("expected BlockLoop, got %v", outerBlock.Type)
	}

	innerBlock, ok := outerBlock.Sections[0].Elements[0].(*Block)
	if !ok {
		t.Fatalf("expected nested Block, got %T", outerBlock.Sections[0].Elements[0])
	}

	if innerBlock.Type != BlockAlt {
		t.Errorf("expected nested BlockAlt, got %v", innerBlock.Type)
	}
}

func TestParseBlockDividerAsFirstContent(t *testing.T) {
	input := `sequenceDiagram
		participant A
		participant B
		alt
		else something
			A->>B: message
		end`

	_, err := Parse(input)
	if err == nil {
		t.Fatal("expected error for divider as first content")
	}
	if !strings.Contains(err.Error(), "divider") || !strings.Contains(err.Error(), "cannot be first content") {
		t.Errorf("expected error about divider as first content, got: %v", err)
	}
}

func TestRenderBlockLoop(t *testing.T) {
	input := `sequenceDiagram
		participant A
		participant B
		loop Every minute
			A->>B: Ping
		end`

	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	output, err := Render(sd, nil)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	if !strings.Contains(output, "loop") {
		t.Errorf("output should contain 'loop':\n%s", output)
	}
	if !strings.Contains(output, "Every minute") {
		t.Errorf("output should contain label:\n%s", output)
	}
	if !strings.Contains(output, "Ping") {
		t.Errorf("output should contain message:\n%s", output)
	}
}

func TestRenderBlockAltElse(t *testing.T) {
	input := `sequenceDiagram
		participant A
		participant B
		alt Success
			A->>B: OK
		else Error
			A->>B: Fail
		end`

	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	output, err := Render(sd, nil)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	if !strings.Contains(output, "alt") {
		t.Errorf("output should contain 'alt':\n%s", output)
	}
	if !strings.Contains(output, "Error") {
		t.Errorf("output should contain 'Error' divider:\n%s", output)
	}
}

func TestRenderBlockNested(t *testing.T) {
	input := `sequenceDiagram
		participant A
		participant B
		loop Retry
			opt Check
				A->>B: Verify
			end
		end`

	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	output, err := Render(sd, nil)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	if !strings.Contains(output, "loop") {
		t.Errorf("output should contain 'loop':\n%s", output)
	}
	if !strings.Contains(output, "opt") {
		t.Errorf("output should contain nested 'opt':\n%s", output)
	}
}

func TestRenderBlockNestedIndentation(t *testing.T) {
	input := `sequenceDiagram
		participant A
		participant B
		loop Outer
			opt Inner
				A->>B: Message
			end
		end`

	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	output, err := Render(sd, nil)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	lines := strings.Split(output, "\n")

	var outerTopLineIdx, innerTopLineIdx int
	outerTopLineIdx, innerTopLineIdx = -1, -1

	for i, line := range lines {
		if strings.Contains(line, "loop Outer") {
			outerTopLineIdx = i
		}
		if strings.Contains(line, "opt Inner") {
			innerTopLineIdx = i
		}
	}

	if outerTopLineIdx == -1 || innerTopLineIdx == -1 {
		t.Fatalf("could not find block header lines in output:\n%s", output)
	}

	isTopLeftCorner := func(ch rune) bool {
		return ch == '┌' || ch == '╭' || ch == '╔' || ch == '╓'
	}
	isTopRightCorner := func(ch rune) bool {
		return ch == '┐' || ch == '╮' || ch == '╗' || ch == '╖'
	}

	findTopLeftCorner := func(lineIdx int) int {
		for i := lineIdx; i >= 0; i-- {
			runes := []rune(lines[i])
			for j, ch := range runes {
				if isTopLeftCorner(ch) {
					return j
				}
			}
		}
		return -1
	}

	findTopRightCorner := func(lineIdx int) int {
		for i := lineIdx; i >= 0; i-- {
			runes := []rune(lines[i])
			for j := len(runes) - 1; j >= 0; j-- {
				if isTopRightCorner(runes[j]) {
					return j
				}
			}
		}
		return -1
	}

	outerLeft := findTopLeftCorner(outerTopLineIdx)
	innerLeft := findTopLeftCorner(innerTopLineIdx)

	outerStartLine := -1
	for i := outerTopLineIdx; i >= 0; i-- {
		runes := []rune(lines[i])
		hasLeft, hasRight := false, false
		for _, ch := range runes {
			if isTopLeftCorner(ch) {
				hasLeft = true
			}
			if isTopRightCorner(ch) {
				hasRight = true
			}
		}
		if hasLeft && hasRight {
			outerStartLine = i
			break
		}
	}
	outerRight := findTopRightCorner(outerStartLine + 1)

	innerStartLine := -1
	for i := innerTopLineIdx; i >= 0; i-- {
		runes := []rune(lines[i])
		count := 0
		for _, ch := range runes {
			if isTopLeftCorner(ch) {
				count++
			}
		}
		if count >= 1 {
			innerStartLine = i
			break
		}
	}
	runes := []rune(lines[innerStartLine])
	innerRight := -1
	for j := len(runes) - 1; j >= 0; j-- {
		if isTopRightCorner(runes[j]) {
			innerRight = j
			break
		}
	}

	if outerLeft == -1 || innerLeft == -1 || outerRight == -1 || innerRight == -1 {
		t.Fatalf("could not find block boundaries (outer: %d-%d, inner: %d-%d) in output:\n%s",
			outerLeft, outerRight, innerLeft, innerRight, output)
	}

	if innerLeft <= outerLeft {
		t.Errorf("inner block left edge (%d) should be greater than outer block left edge (%d) - inner block should be indented inward.\nOutput:\n%s", innerLeft, outerLeft, output)
	}

	if innerRight >= outerRight {
		t.Errorf("inner block right edge (%d) should be less than outer block right edge (%d) - inner block should be contained within outer.\nOutput:\n%s", innerRight, outerRight, output)
	}
}

func TestRenderBlockEmpty(t *testing.T) {
	input := `sequenceDiagram
		participant A
		participant B
		loop Empty
		end`

	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	output, err := Render(sd, nil)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	if !strings.Contains(output, "loop Empty") {
		t.Errorf("output should contain 'loop Empty' label:\n%s", output)
	}
	if !strings.Contains(output, "┌") || !strings.Contains(output, "┐") {
		t.Errorf("output should contain block box corners:\n%s", output)
	}
	if !strings.Contains(output, "└") || !strings.Contains(output, "┘") {
		t.Errorf("output should contain block box bottom corners:\n%s", output)
	}
}

func FuzzParseSequenceDiagram(f *testing.F) {
	f.Add("sequenceDiagram\nA->>B: Hello")
	f.Add("sequenceDiagram\nparticipant Alice\nAlice->>Bob: Hi")
	f.Add("sequenceDiagram\nA-->>B: Response")
	f.Add("sequenceDiagram\nA->>A: Self")

	f.Fuzz(func(t *testing.T, input string) {
		sd, err := Parse(input)
		if err != nil {
			return
		}

		for i, p := range sd.Participants {
			if p.Index != i {
				t.Errorf("Participant %q has incorrect index: got %d, expected %d", p.ID, p.Index, i)
			}
			if p.ID == "" {
				t.Errorf("Participant at index %d has empty ID", i)
			}
			if p.Label == "" {
				t.Errorf("Participant %q has empty label", p.ID)
			}
		}

		for i, msg := range sd.Messages {
			if msg.From == nil || msg.To == nil {
				t.Errorf("Message %d has nil participant", i)
			}
		}

		seen := make(map[string]bool)
		for _, p := range sd.Participants {
			if seen[p.ID] {
				t.Errorf("Duplicate participant ID: %q", p.ID)
			}
			seen[p.ID] = true
		}

		config := diagram.DefaultConfig()
		_, _ = Render(sd, config)
	})
}

func FuzzRenderSequenceDiagram(f *testing.F) {
	seeds := []string{
		"sequenceDiagram\nA->>B: Test",
		"sequenceDiagram\nA->>A: Self",
		"sequenceDiagram\nA->>B: 1\nB->>C: 2\nC->>A: 3",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		sd, err := Parse(input)
		if err != nil {
			return
		}

		for _, useAscii := range []bool{true, false} {
			config := diagram.DefaultConfig()
			config.UseAscii = useAscii

			output, err := Render(sd, config)
			if err != nil {
				return
			}

			if strings.TrimSpace(output) == "" {
				t.Error("Renderer produced empty output for valid diagram")
			}

			for _, p := range sd.Participants {
				if !strings.Contains(output, p.Label) {
					t.Errorf("Rendered output missing participant label: %q", p.Label)
				}
			}

			if !utf8.ValidString(output) {
				t.Error("Rendered output contains invalid UTF-8")
			}
		}
	})
}

func BenchmarkParse(b *testing.B) {
	tests := []struct {
		name         string
		participants int
		messages     int
	}{
		{"small_2p_5m", 2, 5},
		{"medium_5p_20m", 5, 20},
		{"large_10p_50m", 10, 50},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			input := generateDiagram(tt.participants, tt.messages)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := Parse(input)
				if err != nil {
					b.Fatalf("parse failed: %v", err)
				}
			}
		})
	}
}

func BenchmarkRender(b *testing.B) {
	tests := []struct {
		name         string
		participants int
		messages     int
	}{
		{"small_2p_5m", 2, 5},
		{"medium_5p_20m", 5, 20},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			input := generateDiagram(tt.participants, tt.messages)
			sd, err := Parse(input)
			if err != nil {
				b.Fatalf("parse failed: %v", err)
			}
			config := diagram.DefaultConfig()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, renderErr := Render(sd, config)
				if renderErr != nil {
					b.Fatalf("render error: %v", renderErr)
				}
			}
		})
	}
}

func generateDiagram(numParticipants, numMessages int) string {
	var sb strings.Builder
	sb.WriteString("sequenceDiagram\n")
	for i := 0; i < numParticipants; i++ {
		sb.WriteString("    participant P")
		sb.WriteString(string(rune('0' + i)))
		sb.WriteString("\n")
	}
	for i := 0; i < numMessages; i++ {
		from := i % numParticipants
		to := (i + 1) % numParticipants
		arrow := "-"
		if i%2 == 0 {
			arrow = "--"
		}
		sb.WriteString("    P")
		sb.WriteString(string(rune('0' + from)))
		sb.WriteString(arrow)
		sb.WriteString(">>P")
		sb.WriteString(string(rune('0' + to)))
		sb.WriteString(": Message\n")
	}
	return sb.String()
}
