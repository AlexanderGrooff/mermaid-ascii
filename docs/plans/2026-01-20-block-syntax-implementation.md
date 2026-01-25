# Block Syntax Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add support for Mermaid's block syntax (loop, alt, opt, par, critical, break, rect) with arbitrary nesting in sequence diagrams.

**Architecture:** Extend parser with recursive descent for nested blocks, add Block/BlockSection structs implementing DiagramElement, modify renderer to draw block boxes with dividers and handle nesting via depth parameter.

**Tech Stack:** Go, regexp, existing BoxChars for Unicode/ASCII rendering

---

## Task 1: Add Block Data Types

**Files:**
- Modify: `internal/sequence/parser.go`

**Step 1: Add BlockType enum after NotePosition**

```go
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
```

**Step 2: Add Block and BlockSection structs**

```go
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
```

**Step 3: Run build to verify no syntax errors**

Run: `go build ./...`
Expected: Success

**Step 4: Commit**

```bash
git add internal/sequence/parser.go
git commit -m "feat(sequence): add Block data types"
```

---

## Task 2: Add Block Regexes

**Files:**
- Modify: `internal/sequence/parser.go`

**Step 1: Add block regexes to var block**

```go
// blockStartRegex matches block start: loop, alt, opt, par, critical, break, rect
blockStartRegex = regexp.MustCompile(`(?i)^\s*(loop|alt|opt|par|critical|break|rect)\s*(.*)$`)

// blockDividerRegex matches block dividers: else, and, option
blockDividerRegex = regexp.MustCompile(`(?i)^\s*(else|and|option)\s*(.*)$`)

// blockEndRegex matches block end
blockEndRegex = regexp.MustCompile(`(?i)^\s*end\s*$`)
```

**Step 2: Run build to verify no syntax errors**

Run: `go build ./...`
Expected: Success

**Step 3: Commit**

```bash
git add internal/sequence/parser.go
git commit -m "feat(sequence): add block regexes"
```

---

## Task 3: Implement parseBlock Method

**Files:**
- Modify: `internal/sequence/parser.go`

**Step 1: Add helper to map keyword to BlockType**

```go
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
```

**Step 2: Add helper to validate divider for block type**

```go
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
```

**Step 3: Implement parseBlock method**

```go
func (sd *SequenceDiagram) parseBlock(lines []string, startIdx int, participants map[string]*Participant) (*Block, int, error) {
	if startIdx >= len(lines) {
		return nil, startIdx, fmt.Errorf("unexpected end of input")
	}

	// Parse start line
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

		// Check for block end
		if blockEndRegex.MatchString(trimmed) {
			return block, idx + 1, nil
		}

		// Check for divider
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

		// Check for nested block
		if blockStartRegex.MatchString(trimmed) {
			nestedBlock, nextIdx, err := sd.parseBlock(lines, idx, participants)
			if err != nil {
				return nil, idx, fmt.Errorf("nested block: %w", err)
			}
			currentSection.Elements = append(currentSection.Elements, nestedBlock)
			idx = nextIdx
			continue
		}

		// Check for note
		if noteRegex.MatchString(trimmed) {
			if matched, err := sd.parseNote(trimmed, participants); err != nil {
				return nil, idx, err
			} else if matched {
				// Move last element from sd.Elements to currentSection
				if len(sd.Elements) > 0 {
					lastElem := sd.Elements[len(sd.Elements)-1]
					sd.Elements = sd.Elements[:len(sd.Elements)-1]
					currentSection.Elements = append(currentSection.Elements, lastElem)
				}
				idx++
				continue
			}
		}

		// Check for message
		if messageRegex.MatchString(trimmed) {
			if matched, err := sd.parseMessage(trimmed, participants); err != nil {
				return nil, idx, err
			} else if matched {
				// Move last element from sd.Elements to currentSection
				if len(sd.Elements) > 0 {
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

	return nil, idx, fmt.Errorf("block not closed, missing 'end'")
}
```

**Step 4: Run build to verify no syntax errors**

Run: `go build ./...`
Expected: Success

**Step 5: Commit**

```bash
git add internal/sequence/parser.go
git commit -m "feat(sequence): implement parseBlock method"
```

---

## Task 4: Integrate Block Parsing into Main Parse Loop

**Files:**
- Modify: `internal/sequence/parser.go`

**Step 1: Refactor Parse to use line index tracking**

The current Parse function iterates with `for i, line := range lines`. We need to refactor to use index-based iteration so parseBlock can advance the index.

Replace the parse loop with:

```go
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

	sd := &SequenceDiagram{
		Participants: []*Participant{},
		Messages:     []*Message{},
		Elements:     []DiagramElement{},
		Autonumber:   false,
	}
	participantMap := make(map[string]*Participant)

	idx := 1
	for idx < len(lines) {
		line := lines[idx]
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			idx++
			continue
		}

		// Check for autonumber
		if autonumberRegex.MatchString(trimmed) {
			sd.Autonumber = true
			idx++
			continue
		}

		// Check for participant
		if matched, err := sd.parseParticipant(trimmed, participantMap); err != nil {
			return nil, fmt.Errorf("line %d: %w", idx+1, err)
		} else if matched {
			idx++
			continue
		}

		// Check for block start
		if blockStartRegex.MatchString(trimmed) {
			block, nextIdx, err := sd.parseBlock(lines, idx, participantMap)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", idx+1, err)
			}
			sd.Elements = append(sd.Elements, block)
			idx = nextIdx
			continue
		}

		// Check for message
		if matched, err := sd.parseMessage(trimmed, participantMap); err != nil {
			return nil, fmt.Errorf("line %d: %w", idx+1, err)
		} else if matched {
			idx++
			continue
		}

		// Check for note
		if matched, err := sd.parseNote(trimmed, participantMap); err != nil {
			return nil, fmt.Errorf("line %d: %w", idx+1, err)
		} else if matched {
			idx++
			continue
		}

		return nil, fmt.Errorf("line %d: invalid syntax: %q", idx+1, trimmed)
	}

	if len(sd.Participants) == 0 {
		return nil, fmt.Errorf("no participants found")
	}

	return sd, nil
}
```

**Step 2: Run existing tests to verify no regression**

Run: `go test ./internal/sequence/... -v`
Expected: All existing tests pass

**Step 3: Commit**

```bash
git add internal/sequence/parser.go
git commit -m "feat(sequence): integrate block parsing into main loop"
```

---

## Task 5: Write Parser Tests for Blocks

**Files:**
- Modify: `internal/sequence/sequence_test.go`

**Step 1: Write test for simple loop block**

```go
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
```

**Step 2: Write test for alt/else block**

```go
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
```

**Step 3: Write test for par/and block**

```go
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
```

**Step 4: Write test for nested blocks**

```go
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
```

**Step 5: Run all block parser tests**

Run: `go test ./internal/sequence/... -run TestParseBlock -v`
Expected: All PASS

**Step 6: Commit**

```bash
git add internal/sequence/sequence_test.go
git commit -m "test(sequence): add parser tests for block syntax"
```

---

## Task 6: Add renderBlock Stub and Update Render Loop

**Files:**
- Modify: `internal/sequence/renderer.go`

**Step 1: Add renderBlock stub**

```go
func renderBlock(block *Block, layout *diagramLayout, chars BoxChars, depth int) []string {
	// TODO: implement
	return nil
}
```

**Step 2: Add findBlockParticipantRange helper**

```go
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
				if nestedMin != -1 && (minIdx == -1 || nestedMin < minIdx) {
					minIdx = nestedMin
				}
				if nestedMax > maxIdx {
					maxIdx = nestedMax
				}
			}
		}
	}

	for _, section := range block.Sections {
		findInElements(section.Elements)
	}

	return minIdx, maxIdx
}
```

**Step 3: Update Render loop to handle blocks**

In the element type switch, add case for Block:

```go
case *Block:
	blockLines := renderBlock(e, layout, chars, 0)
	if blockLines != nil {
		lines = append(lines, blockLines...)
	}
```

**Step 4: Run existing tests to verify no regression**

Run: `go test ./internal/sequence/... -v`
Expected: All existing tests pass

**Step 5: Commit**

```bash
git add internal/sequence/renderer.go
git commit -m "feat(sequence): add renderBlock stub and update render loop"
```

---

## Task 7: Implement renderBlock

**Files:**
- Modify: `internal/sequence/renderer.go`

**Step 1: Write test for block rendering**

```go
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/sequence/... -run TestRenderBlockLoop -v`
Expected: FAIL

**Step 3: Implement renderBlock**

```go
func renderBlock(block *Block, layout *diagramLayout, chars BoxChars, depth int) []string {
	var lines []string

	minIdx, maxIdx := findBlockParticipantRange(block)
	if minIdx == -1 || maxIdx == -1 {
		return nil
	}

	indent := depth * 2
	leftCenter := layout.participantCenters[minIdx]
	rightCenter := layout.participantCenters[maxIdx]

	boxLeft := leftCenter - 3 - indent
	if boxLeft < 0 {
		boxLeft = 0
	}
	boxRight := rightCenter + 3

	// Ensure minimum width for label
	headerLabel := fmt.Sprintf("%s %s", block.Type, block.Label)
	labelWidth := runewidth.StringWidth(headerLabel)
	if boxRight-boxLeft < labelWidth+4 {
		boxRight = boxLeft + labelWidth + 4
	}

	ensureWidth := boxRight + 1
	if ensureWidth < layout.totalWidth {
		ensureWidth = layout.totalWidth
	}

	// Helper to create a line with lifelines
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

	// Draw top border
	topLine := makeLine()
	topLine[boxLeft] = chars.TopLeft
	for i := boxLeft + 1; i < boxRight; i++ {
		if topLine[i] == chars.Vertical {
			topLine[i] = chars.TeeUp
		} else {
			topLine[i] = chars.Horizontal
		}
	}
	topLine[boxRight] = chars.TopRight
	lines = append(lines, strings.TrimRight(string(topLine), " "))

	// Draw header label line
	headerLine := makeLine()
	headerLine[boxLeft] = chars.Vertical
	headerLine[boxRight] = chars.Vertical
	col := boxLeft + 2
	for _, r := range headerLabel {
		if col < boxRight {
			headerLine[col] = r
			col++
		}
	}
	lines = append(lines, strings.TrimRight(string(headerLine), " "))

	// Draw header separator
	sepLine := makeLine()
	sepLine[boxLeft] = chars.Vertical
	for i := boxLeft + 1; i < boxRight; i++ {
		if sepLine[i] == chars.Vertical {
			// keep lifeline
		} else {
			sepLine[i] = chars.Horizontal
		}
	}
	sepLine[boxRight] = chars.Vertical
	lines = append(lines, strings.TrimRight(string(sepLine), " "))

	// Render each section
	for sectionIdx, section := range block.Sections {
		// Draw section divider if not first section
		if sectionIdx > 0 {
			divLine := makeLine()
			divLine[boxLeft] = chars.TeeRight
			for i := boxLeft + 1; i < boxRight; i++ {
				if divLine[i] == chars.Vertical {
					divLine[i] = chars.Cross
				} else {
					divLine[i] = chars.Horizontal
				}
			}
			divLine[boxRight] = chars.TeeLeft
			lines = append(lines, strings.TrimRight(string(divLine), " "))

			// Section label
			if section.Label != "" {
				labelLine := makeLine()
				labelLine[boxLeft] = chars.Vertical
				labelLine[boxRight] = chars.Vertical
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

		// Render section elements
		for _, elem := range section.Elements {
			// Add spacing line
			spaceLine := makeLine()
			spaceLine[boxLeft] = chars.Vertical
			spaceLine[boxRight] = chars.Vertical
			lines = append(lines, strings.TrimRight(string(spaceLine), " "))

			switch e := elem.(type) {
			case *Message:
				msgLines := renderMessage(e, layout, chars)
				for _, ml := range msgLines {
					// Add block borders to message lines
					mlRunes := []rune(ml)
					for len(mlRunes) <= ensureWidth {
						mlRunes = append(mlRunes, ' ')
					}
					mlRunes[boxLeft] = chars.Vertical
					mlRunes[boxRight] = chars.Vertical
					lines = append(lines, strings.TrimRight(string(mlRunes), " "))
				}
			case *Note:
				noteLines := renderNote(e, layout, chars)
				for _, nl := range noteLines {
					nlRunes := []rune(nl)
					for len(nlRunes) <= ensureWidth {
						nlRunes = append(nlRunes, ' ')
					}
					nlRunes[boxLeft] = chars.Vertical
					nlRunes[boxRight] = chars.Vertical
					lines = append(lines, strings.TrimRight(string(nlRunes), " "))
				}
			case *Block:
				nestedLines := renderBlock(e, layout, chars, depth+1)
				for _, nl := range nestedLines {
					nlRunes := []rune(nl)
					for len(nlRunes) <= ensureWidth {
						nlRunes = append(nlRunes, ' ')
					}
					nlRunes[boxLeft] = chars.Vertical
					nlRunes[boxRight] = chars.Vertical
					lines = append(lines, strings.TrimRight(string(nlRunes), " "))
				}
			}
		}
	}

	// Add spacing before bottom
	spaceLine := makeLine()
	spaceLine[boxLeft] = chars.Vertical
	spaceLine[boxRight] = chars.Vertical
	lines = append(lines, strings.TrimRight(string(spaceLine), " "))

	// Draw bottom border
	bottomLine := makeLine()
	bottomLine[boxLeft] = chars.BottomLeft
	for i := boxLeft + 1; i < boxRight; i++ {
		if bottomLine[i] == chars.Vertical {
			bottomLine[i] = chars.TeeDown
		} else {
			bottomLine[i] = chars.Horizontal
		}
	}
	bottomLine[boxRight] = chars.BottomRight
	lines = append(lines, strings.TrimRight(string(bottomLine), " "))

	return lines
}
```

**Step 4: Add Cross character to BoxChars if missing**

Check `internal/sequence/charset.go` and add `Cross` character ('+' for ASCII, '┼' for Unicode) if not present.

**Step 5: Run test to verify it passes**

Run: `go test ./internal/sequence/... -run TestRenderBlockLoop -v`
Expected: PASS

**Step 6: Commit**

```bash
git add internal/sequence/renderer.go internal/sequence/charset.go
git commit -m "feat(sequence): implement renderBlock"
```

---

## Task 8: Write Renderer Tests for Multi-Section and Nested Blocks

**Files:**
- Modify: `internal/sequence/sequence_test.go`

**Step 1: Write test for alt/else rendering**

```go
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
```

**Step 2: Write test for nested blocks rendering**

```go
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
```

**Step 3: Run all render tests**

Run: `go test ./internal/sequence/... -run TestRenderBlock -v`
Expected: All PASS

**Step 4: Commit**

```bash
git add internal/sequence/sequence_test.go
git commit -m "test(sequence): add renderer tests for multi-section and nested blocks"
```

---

## Task 9: Integration Test with scratch/note-syntax.mermaid

**Step 1: Run CLI against the scratch file**

Run: `cat scratch/note-syntax.mermaid | go run main.go`

This file contains:
- `loop Video Stream` with messages
- `par Control Commands` with `and Events`

Verify the output shows both blocks rendered correctly with their contents.

**Step 2: If errors, debug and fix**

Common issues:
- Parser errors (block not recognized)
- Missing 'end' keyword
- Incorrect divider validation

**Step 3: Run full test suite**

Run: `go test ./... -v`
Expected: All PASS

**Step 4: Commit if any fixes were needed**

```bash
git add -A
git commit -m "fix(sequence): address integration test issues"
```

---

## Task 10: Final Cleanup and Documentation

**Step 1: Run all tests**

Run: `go test ./... -v`
Expected: All PASS

**Step 2: Test visual output with comprehensive example**

```bash
cat << 'EOF' | go run main.go
sequenceDiagram
    participant A
    participant B
    participant C
    A->>B: Request
    loop Retry 3 times
        B->>C: Check
        alt Success
            C-->>B: OK
        else Failure
            C-->>B: Error
        end
    end
    B-->>A: Response
EOF
```

**Step 3: Final commit**

```bash
git add -A
git commit -m "feat(sequence): complete block syntax support (loop, alt, opt, par, critical, break, rect)"
```

---

## Summary

After completing all tasks, the sequence diagram parser will support:
- `loop [label]` ... `end`
- `alt [label]` ... `else [label]` ... `end`
- `opt [label]` ... `end`
- `par [label]` ... `and [label]` ... `end`
- `critical [label]` ... `option [label]` ... `end`
- `break [label]` ... `end`
- `rect [color]` ... `end`
- Arbitrary nesting of blocks
- Blocks mixed with messages and notes
