# Note Syntax Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add support for Mermaid's `Note over`, `Note left of`, and `Note right of` syntax in sequence diagrams.

**Architecture:** Extend parser with note regex and parseNote method, add Note struct and DiagramElement interface, modify renderer to iterate Elements and dispatch to renderNote functions.

**Tech Stack:** Go, regexp, existing BoxChars for Unicode/ASCII rendering

---

## Task 1: Add Note Data Types

**Files:**
- Modify: `internal/sequence/parser.go:28-48`

**Step 1: Add NotePosition enum and Note struct after ArrowType**

Add after line 54 in `internal/sequence/parser.go`:

```go
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
```

**Step 2: Add DiagramElement interface and Elements field**

Add interface before SequenceDiagram struct:

```go
type DiagramElement interface {
	isElement()
}

func (*Message) isElement() {}
func (*Note) isElement()    {}
```

Modify SequenceDiagram struct to add Elements field:

```go
type SequenceDiagram struct {
	Participants []*Participant
	Messages     []*Message
	Elements     []DiagramElement // Ordered messages and notes
	Autonumber   bool
}
```

**Step 3: Run build to verify no syntax errors**

Run: `go build ./...`
Expected: Success, no errors

**Step 4: Commit**

```bash
git add internal/sequence/parser.go
git commit -m "feat(sequence): add Note data types and DiagramElement interface"
```

---

## Task 2: Add Note Regex and Parser Method

**Files:**
- Modify: `internal/sequence/parser.go:17-26` (add regex)
- Modify: `internal/sequence/parser.go` (add parseNote method)

**Step 1: Add noteRegex to var block**

Add to the var block after autonumberRegex:

```go
// noteRegex matches note declarations:
//   Note over Actor: text
//   Note over Actor1,Actor2: text
//   Note left of Actor: text
//   Note right of Actor: text
noteRegex = regexp.MustCompile(`(?i)^\s*note\s+(over|left\s+of|right\s+of)\s+([^:]+):\s*(.*)$`)
```

**Step 2: Add parseNote method**

Add after parseMessage method:

```go
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
```

**Step 3: Run build to verify no syntax errors**

Run: `go build ./...`
Expected: Success, no errors

**Step 4: Commit**

```bash
git add internal/sequence/parser.go
git commit -m "feat(sequence): add noteRegex and parseNote method"
```

---

## Task 3: Integrate Note Parsing into Main Parse Loop

**Files:**
- Modify: `internal/sequence/parser.go:103-128` (parse loop)
- Modify: `internal/sequence/parser.go:167-208` (parseMessage to also append to Elements)

**Step 1: Update parseMessage to append to Elements**

In parseMessage, before `return true, nil`, add:

```go
sd.Elements = append(sd.Elements, msg)
```

**Step 2: Add parseNote call in main parse loop**

In the Parse function, after the parseMessage block and before the error return, add:

```go
if matched, err := sd.parseNote(trimmed, participantMap); err != nil {
	return nil, fmt.Errorf("line %d: %w", i+2, err)
} else if matched {
	continue
}
```

**Step 3: Run build to verify no syntax errors**

Run: `go build ./...`
Expected: Success, no errors

**Step 4: Commit**

```bash
git add internal/sequence/parser.go
git commit -m "feat(sequence): integrate note parsing into main loop"
```

---

## Task 4: Write Parser Tests for Notes

**Files:**
- Modify: `internal/sequence/sequence_test.go`

**Step 1: Write test for Note over single actor**

Add test function:

```go
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
		t.Errorf("expected actor A, got %v", note.Actors)
	}
	if note.Text != "This is a note" {
		t.Errorf("expected 'This is a note', got %q", note.Text)
	}
}
```

**Step 2: Run test to verify it passes**

Run: `go test ./internal/sequence/... -run TestParseNoteOverSingleActor -v`
Expected: PASS

**Step 3: Write test for Note over multiple actors**

```go
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
		t.Errorf("expected 2 actors, got %d", len(note.Actors))
	}
	if note.Actors[0].ID != "A" || note.Actors[1].ID != "B" {
		t.Errorf("expected actors A and B, got %v", note.Actors)
	}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/sequence/... -run TestParseNoteOverMultipleActors -v`
Expected: PASS

**Step 5: Write test for Note left of and right of**

```go
func TestParseNoteLeftRight(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		position NotePosition
	}{
		{
			name: "left of",
			input: `sequenceDiagram
				participant A
				Note left of A: Left note`,
			position: NoteLeftOf,
		},
		{
			name: "right of",
			input: `sequenceDiagram
				participant A
				Note right of A: Right note`,
			position: NoteRightOf,
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

			if note.Position != tt.position {
				t.Errorf("expected %v, got %v", tt.position, note.Position)
			}
		})
	}
}
```

**Step 6: Run all note tests**

Run: `go test ./internal/sequence/... -run TestParseNote -v`
Expected: All PASS

**Step 7: Commit**

```bash
git add internal/sequence/sequence_test.go
git commit -m "test(sequence): add parser tests for note syntax"
```

---

## Task 5: Add renderNote Stub and Update Render Loop

**Files:**
- Modify: `internal/sequence/renderer.go:115-128` (render loop)

**Step 1: Add renderNote stub function**

Add after renderSelfMessage function:

```go
func renderNote(note *Note, layout *diagramLayout, chars BoxChars) []string {
	// TODO: implement note rendering
	return nil
}
```

**Step 2: Update Render function to iterate Elements**

Replace the message loop (lines ~115-125) with:

```go
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
	}
}
```

**Step 3: Run existing tests to verify no regression**

Run: `go test ./internal/sequence/... -v`
Expected: All existing tests PASS

**Step 4: Commit**

```bash
git add internal/sequence/renderer.go
git commit -m "refactor(sequence): update render loop to iterate Elements"
```

---

## Task 6: Implement renderNoteOver

**Files:**
- Modify: `internal/sequence/renderer.go`

**Step 1: Write test for Note over rendering**

Add to test file:

```go
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

	// Verify note box appears in output
	if !strings.Contains(output, "Test note") {
		t.Errorf("output should contain note text:\n%s", output)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/sequence/... -run TestRenderNoteOver -v`
Expected: FAIL (note text not in output)

**Step 3: Implement renderNoteOver**

Add helper and implementation:

```go
func renderNoteOver(note *Note, layout *diagramLayout, chars BoxChars) []string {
	var lines []string

	// Calculate horizontal bounds
	leftActor := note.Actors[0]
	rightActor := note.Actors[len(note.Actors)-1]
	
	leftCenter := layout.participantCenters[leftActor.Index]
	rightCenter := layout.participantCenters[rightActor.Index]
	
	if leftCenter > rightCenter {
		leftCenter, rightCenter = rightCenter, leftCenter
	}

	// Note box width: from left actor center to right actor center, with padding
	padding := 2
	boxLeft := leftCenter - padding
	if boxLeft < 0 {
		boxLeft = 0
	}
	textWidth := runewidth.StringWidth(note.Text)
	minBoxWidth := textWidth + 4 // 2 padding each side
	boxWidth := rightCenter - leftCenter + padding*2
	if boxWidth < minBoxWidth {
		boxWidth = minBoxWidth
	}
	boxRight := boxLeft + boxWidth

	// Build top border with lifeline connections
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

	// Build text line
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
	// Center text in box
	textStart := boxLeft + (boxWidth-textWidth)/2
	col := textStart
	for _, r := range note.Text {
		if col < len(textLine) && col < boxRight {
			textLine[col] = r
			col++
		}
	}
	lines = append(lines, strings.TrimRight(string(textLine), " "))

	// Build bottom border with lifeline connections
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
```

Update renderNote to call it:

```go
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
```

Add stubs for left/right:

```go
func renderNoteLeftOf(note *Note, layout *diagramLayout, chars BoxChars) []string {
	// TODO: implement
	return nil
}

func renderNoteRightOf(note *Note, layout *diagramLayout, chars BoxChars) []string {
	// TODO: implement
	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/sequence/... -run TestRenderNoteOver -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/sequence/renderer.go internal/sequence/sequence_test.go
git commit -m "feat(sequence): implement renderNoteOver"
```

---

## Task 7: Implement renderNoteLeftOf

**Files:**
- Modify: `internal/sequence/renderer.go`
- Modify: `internal/sequence/sequence_test.go`

**Step 1: Write test for Note left of rendering**

```go
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/sequence/... -run TestRenderNoteLeftOf -v`
Expected: FAIL

**Step 3: Implement renderNoteLeftOf**

```go
func renderNoteLeftOf(note *Note, layout *diagramLayout, chars BoxChars) []string {
	var lines []string
	actor := note.Actors[0]
	center := layout.participantCenters[actor.Index]

	textWidth := runewidth.StringWidth(note.Text)
	boxWidth := textWidth + 4 // padding
	boxRight := center - 2
	boxLeft := boxRight - boxWidth
	if boxLeft < 0 {
		boxLeft = 0
		boxWidth = boxRight - boxLeft
	}

	// Top border
	topLine := make([]rune, layout.totalWidth+1)
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
	textLine := make([]rune, layout.totalWidth+1)
	for i := range textLine {
		textLine[i] = ' '
	}
	for _, c := range layout.participantCenters {
		if c < len(textLine) {
			textLine[c] = chars.Vertical
		}
	}
	textLine[boxLeft] = chars.Vertical
	textLine[boxRight] = chars.TeeLeft
	for i := boxRight + 1; i < center; i++ {
		textLine[i] = chars.Horizontal
	}
	textLine[center] = chars.TeeLeft
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
	bottomLine := make([]rune, layout.totalWidth+1)
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
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/sequence/... -run TestRenderNoteLeftOf -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/sequence/renderer.go internal/sequence/sequence_test.go
git commit -m "feat(sequence): implement renderNoteLeftOf"
```

---

## Task 8: Implement renderNoteRightOf

**Files:**
- Modify: `internal/sequence/renderer.go`
- Modify: `internal/sequence/sequence_test.go`

**Step 1: Write test for Note right of rendering**

```go
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/sequence/... -run TestRenderNoteRightOf -v`
Expected: FAIL

**Step 3: Implement renderNoteRightOf**

```go
func renderNoteRightOf(note *Note, layout *diagramLayout, chars BoxChars) []string {
	var lines []string
	actor := note.Actors[0]
	center := layout.participantCenters[actor.Index]

	textWidth := runewidth.StringWidth(note.Text)
	boxWidth := textWidth + 4
	boxLeft := center + 2
	boxRight := boxLeft + boxWidth

	ensureWidth := layout.totalWidth
	if boxRight > ensureWidth {
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
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/sequence/... -run TestRenderNoteRightOf -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/sequence/renderer.go internal/sequence/sequence_test.go
git commit -m "feat(sequence): implement renderNoteRightOf"
```

---

## Task 9: Integration Test with Real-World Example

**Files:**
- Modify: `internal/sequence/sequence_test.go`

**Step 1: Write integration test using scratch file syntax**

```go
func TestRenderNoteIntegration(t *testing.T) {
	input := `sequenceDiagram
		participant Client
		participant Server
		Client->>Server: Request
		Note over Client: Processing locally
		Server-->>Client: Response
		Note over Client,Server: Transaction complete`

	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(sd.Elements) != 4 {
		t.Fatalf("expected 4 elements (2 messages + 2 notes), got %d", len(sd.Elements))
	}

	output, err := Render(sd, nil)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	// Verify all elements rendered
	if !strings.Contains(output, "Request") {
		t.Error("missing Request message")
	}
	if !strings.Contains(output, "Processing locally") {
		t.Error("missing first note")
	}
	if !strings.Contains(output, "Response") {
		t.Error("missing Response message")
	}
	if !strings.Contains(output, "Transaction complete") {
		t.Error("missing second note")
	}

	t.Logf("Rendered output:\n%s", output)
}
```

**Step 2: Run integration test**

Run: `go test ./internal/sequence/... -run TestRenderNoteIntegration -v`
Expected: PASS

**Step 3: Run all tests to verify no regressions**

Run: `go test ./internal/sequence/... -v`
Expected: All PASS

**Step 4: Commit**

```bash
git add internal/sequence/sequence_test.go
git commit -m "test(sequence): add note integration test"
```

---

## Task 10: Test with scratch/note-syntax.mermaid

**Step 1: Run CLI against the scratch file**

Run: `go run main.go < scratch/note-syntax.mermaid`

Verify the note on line 10 (`Note over Client: Extract control server from connectToken`) renders correctly.

**Step 2: If errors, debug and fix**

Check for:
- Parser errors (note not recognized)
- Renderer errors (layout issues)
- Visual issues (note box misaligned)

**Step 3: Final test run**

Run: `go test ./... -v`
Expected: All PASS

**Step 4: Final commit**

```bash
git add -A
git commit -m "feat(sequence): complete note syntax support"
```

---

## Summary

After completing all tasks, the sequence diagram parser will support:
- `Note over Actor: text`
- `Note over Actor1,Actor2: text`
- `Note left of Actor: text`
- `Note right of Actor: text`

All existing functionality remains backward compatible.
