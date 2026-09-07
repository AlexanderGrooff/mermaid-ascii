package sequence

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/AlexanderGrooff/mermaid-ascii/pkg/diagram"
	"github.com/AlexanderGrooff/mermaid-ascii/pkg/diagram/testutil"
	"github.com/mattn/go-runewidth"
)

// getTestDataPath returns the absolute path to testdata directory.
// This avoids brittle relative paths by using runtime.Caller to find the test file location.
func getTestDataPath() string {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	return filepath.Join(dir, "..", "..", "cmd", "testdata")
}

// TestSequenceDiagramRendering tests all sequence diagram test cases with Unicode charset.
func TestSequenceDiagramRendering(t *testing.T) {
	testDataPath := filepath.Join(getTestDataPath(), "sequence")

	// Test files - stored in sequence/ directory (Unicode expected output)
	testFiles := []string{
		"adjacent_participants_communication.txt",
		"alt_basic.txt",
		"alt_multiple_else.txt",
		"par_basic.txt",
		"critical_basic.txt",
		"break_rect.txt",
		"arrow_types.txt",
		"note_over_spaced_names.txt",
		"fragment_spaced_names.txt",
		"self_message_spaced_name.txt",
		"quoted_name_with_arrow.txt",
		"central_spaced_names.txt",
		"participant_names_spaces.txt",
		"participant_names_dashes.txt",
		"participant_names_alias_equals.txt",
		"actor_alias.txt",
		"actor_participant_mix.txt",
		"box_basic.txt",
		"box_multiple.txt",
		"activation_shorthand.txt",
		"activation_stacked.txt",
		"activation_keywords.txt",
		"create_destroy.txt",
		"create_actor_alias.txt",
		"central_connections.txt",
		"central_connection_self.txt",
		"cross_arrows.txt",
		"async_point_arrows.txt",
		"bidirectional_arrows.txt",
		"self_arrow_variants.txt",
		"autonumber.txt",
		"bidirectional_messages.txt",
		"dotted_arrows_only.txt",
		"four_participants.txt",
		"fragment_partial_span.txt",
		"long_participant_names.txt",
		"loop_autonumber.txt",
		"loop_basic.txt",
		"loop_empty.txt",
		"messages_without_labels.txt",
		"multiword_labels.txt",
		"note_over_single.txt",
		"note_over_span.txt",
		"note_in_loop.txt",
		"opt_basic.txt",
		"opt_no_label.txt",
		"self_message.txt",
		"simple_two_participants.txt",
		"single_message.txt",
		"three_participants.txt",
		"east_asian_participants.txt",
		"mixed_width_cjk.txt",
		"cjk_matrix.txt",
		"combining_marks_mixed_width.txt",
		"leading_combining_marks.txt",
		"literal_control_character.txt",
	}

	for _, testFile := range testFiles {
		t.Run(testFile, func(t *testing.T) {
			verifySequenceDiagramWithCharset(t, filepath.Join(testDataPath, testFile), false)
		})
	}
}

// TestSequenceDiagramRendering_ASCII tests ASCII rendering with golden files.
// These golden files ensure ASCII output remains correct and readable.
func TestSequenceDiagramRendering_ASCII(t *testing.T) {
	testDataPath := filepath.Join(getTestDataPath(), "sequence-ascii")

	goldenFiles := []string{
		"alt_basic.txt",
		"par_basic.txt",
		"note_over_single.txt",
		"arrow_types.txt",
		"participant_names_alias_equals.txt",
		"note_over_spaced_names.txt",
		"fragment_spaced_names.txt",
		"participant_names_spaces.txt",
		"participant_names_dashes.txt",
		"actor_alias.txt",
		"actor_participant_mix.txt",
		"box_basic.txt",
		"box_multiple.txt",
		"activation_shorthand.txt",
		"activation_stacked.txt",
		"activation_keywords.txt",
		"create_destroy.txt",
		"create_actor_alias.txt",
		"central_connections.txt",
		"central_connection_self.txt",
		"cross_arrows.txt",
		"async_point_arrows.txt",
		"bidirectional_arrows.txt",
		"self_arrow_variants.txt",
		"autonumber.txt",
		"dotted_arrows_only.txt",
		"loop_basic.txt",
		"loop_empty.txt",
		"opt_basic.txt",
		"self_message.txt",
		"simple_two_participants.txt",
		"three_participants.txt",
		"east_asian_participants.txt",
		"mixed_width_cjk.txt",
		"cjk_matrix.txt",
		"combining_marks_mixed_width.txt",
		"leading_combining_marks.txt",
	}

	for _, testFile := range goldenFiles {
		t.Run(testFile, func(t *testing.T) {
			verifySequenceDiagramWithCharset(t, filepath.Join(testDataPath, testFile), true)
		})
	}
}

// TestSequenceDiagramRendering_ASCIISmokeTest verifies ASCII rendering works for all test inputs.
// This ensures parsing and ASCII rendering don't crash, without requiring golden files for every case.
func TestSequenceDiagramRendering_ASCIISmokeTest(t *testing.T) {
	testDataPath := filepath.Join(getTestDataPath(), "sequence")

	testFiles := []string{
		"adjacent_participants_communication.txt",
		"alt_basic.txt",
		"alt_multiple_else.txt",
		"par_basic.txt",
		"critical_basic.txt",
		"break_rect.txt",
		"arrow_types.txt",
		"note_over_spaced_names.txt",
		"fragment_spaced_names.txt",
		"self_message_spaced_name.txt",
		"quoted_name_with_arrow.txt",
		"central_spaced_names.txt",
		"participant_names_spaces.txt",
		"participant_names_dashes.txt",
		"participant_names_alias_equals.txt",
		"central_connections.txt",
		"central_connection_self.txt",
		"cross_arrows.txt",
		"async_point_arrows.txt",
		"bidirectional_arrows.txt",
		"self_arrow_variants.txt",
		"autonumber.txt",
		"bidirectional_messages.txt",
		"dotted_arrows_only.txt",
		"four_participants.txt",
		"fragment_partial_span.txt",
		"long_participant_names.txt",
		"loop_autonumber.txt",
		"loop_basic.txt",
		"loop_empty.txt",
		"messages_without_labels.txt",
		"multiword_labels.txt",
		"note_over_single.txt",
		"note_over_span.txt",
		"note_in_loop.txt",
		"opt_basic.txt",
		"opt_no_label.txt",
		"self_message.txt",
		"simple_two_participants.txt",
		"single_message.txt",
		"three_participants.txt",
		"east_asian_participants.txt",
	}

	for _, testFile := range testFiles {
		t.Run(testFile, func(t *testing.T) {
			tc, err := testutil.ReadSequenceTestCase(filepath.Join(testDataPath, testFile))
			if err != nil {
				t.Fatalf("Failed to read test case: %v", err)
			}

			sd, err := Parse(tc.Mermaid)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			config := diagram.NewTestConfig(true, "cli") // ASCII, CLI style

			output, err := Render(sd, config)
			if err != nil {
				t.Fatalf("Failed to render ASCII: %v", err)
			}

			// Smoke test - just ensure output is not empty and contains expected elements
			if len(output) == 0 {
				t.Error("ASCII output is empty")
			}

			// Verify all participant labels appear in output
			for _, p := range sd.Participants {
				if !strings.Contains(output, p.Label) {
					t.Errorf("ASCII output missing participant label: %q", p.Label)
				}
			}

			// Verify ASCII characters are used (not Unicode box-drawing)
			if strings.ContainsAny(output, "┌┐└┘├┤┬┴┼│─►◄┈") {
				t.Error("ASCII output contains Unicode box-drawing characters")
			}
		})
	}
}

// TestSequenceDiagramRendering_EastAsian tests Unicode rendering with East Asian locale settings.
// This ensures proper spacing between participant boxes when runewidth treats
// box-drawing characters as double-width (EastAsianWidth=true).
func TestSequenceDiagramRendering_EastAsian(t *testing.T) {
	// Save original condition
	origCond := runewidth.DefaultCondition

	// Set East Asian width mode directly (more reliable than environment variable)
	eastAsianCond := runewidth.NewCondition()
	eastAsianCond.EastAsianWidth = true
	runewidth.DefaultCondition = eastAsianCond

	// Restore original settings after test
	defer func() {
		runewidth.DefaultCondition = origCond
	}()

	testDataPath := filepath.Join(getTestDataPath(), "sequence")

	testFiles := []string{
		"simple_two_participants.txt",
		"three_participants.txt",
		"east_asian_participants.txt",
		"four_participants.txt",
		"leading_combining_marks.txt",
		"literal_control_character.txt",
	}

	for _, testFile := range testFiles {
		t.Run(testFile, func(t *testing.T) {
			verifySequenceDiagramWithCharset(t, filepath.Join(testDataPath, testFile), false)
		})
	}
}

func TestCJKFeatureMatrixKeepsTextAcrossCharsets(t *testing.T) {
	input := `sequenceDiagram
box 日本語サービス
participant A as 顧客
participant B as 服务
end
actor C as 監査
participant D as 数据库
A->>B: 注文
B-->>A: 応答
B->>B: 自己
Note over A,B: 注記
alt 成功
 A-xC: 失敗
else 失敗
 C-)D: 完了
end
loop 再試行
 D<<->>A: 再送
end
activate B
B-->>D: 更新
deactivate B
create participant E as 新規
D->>E: 作成
destroy E
A-xE: 終了`

	var unicodeOutput, asciiOutput string
	for _, useASCII := range []bool{false, true} {
		d, err := Parse(input)
		if err != nil {
			t.Fatalf("Parse(useASCII=%t): %v", useASCII, err)
		}
		config := diagram.NewTestConfig(useASCII, "cli")
		output, err := Render(d, config)
		if err != nil {
			t.Fatalf("Render(useASCII=%t): %v", useASCII, err)
		}
		for _, text := range []string{
			"日本語サービス", "顧客", "服务", "監査", "数据库", "注文", "応答", "自己",
			"注記", "成功", "失敗", "完了", "再試行", "再送", "更新", "新規", "作成", "終了",
		} {
			if !strings.Contains(output, text) {
				t.Errorf("Render(useASCII=%t) missing CJK text %q:\n%s", useASCII, text, output)
			}
		}
		assertCJKFeatureAlignment(t, output, useASCII)
		if useASCII {
			asciiOutput = output
		} else {
			unicodeOutput = output
		}
	}
	assertCJKFeatureRowsHaveConsistentWidth(t, unicodeOutput, asciiOutput)
}

func assertCJKFeatureRowsHaveConsistentWidth(t *testing.T, unicodeOutput, asciiOutput string) {
	t.Helper()
	unicodeRows := strings.Split(strings.TrimRight(unicodeOutput, "\n"), "\n")
	asciiRows := strings.Split(strings.TrimRight(asciiOutput, "\n"), "\n")
	if len(unicodeRows) != len(asciiRows) {
		t.Fatalf("Unicode and ASCII row counts differ: %d vs %d", len(unicodeRows), len(asciiRows))
	}
	for i := range unicodeRows {
		unicodeWidth := displayWidth(unicodeRows[i])
		asciiWidth := displayWidth(asciiRows[i])
		if unicodeWidth != len(textCells(unicodeRows[i])) || asciiWidth != len(textCells(asciiRows[i])) {
			t.Errorf("row %d display width does not match text cells: Unicode=%d ASCII=%d", i, unicodeWidth, asciiWidth)
		}
		if unicodeWidth != asciiWidth {
			t.Errorf("row %d display widths differ: Unicode=%d ASCII=%d", i, unicodeWidth, asciiWidth)
		}
	}
}

func assertCJKFeatureAlignment(t *testing.T, output string, useASCII bool) {
	t.Helper()
	lines := strings.Split(strings.TrimSuffix(output, "\n"), "\n")
	labels := []string{"顧客", "服务", "監査", "数据库", "新規"}
	columns := participantColumns(t, lines, labels, useASCII)

	for _, message := range []struct {
		label, from, to string
	}{
		{"注文", "顧客", "服务"},
		{"応答", "服务", "顧客"},
		{"失敗", "顧客", "監査"},
		{"完了", "監査", "数据库"},
		{"再送", "数据库", "顧客"},
		{"更新", "服务", "数据库"},
		{"作成", "数据库", "新規"},
		{"終了", "顧客", "新規"},
	} {
		assertAlignedMessage(t, lines, message.label, columns[message.from], columns[message.to], useASCII)
	}
	assertAlignedSelfMessage(t, lines, "自己", columns["服务"])
	assertAlignedNote(t, lines, "注記", columns["顧客"], columns["服务"], useASCII)
	assertAlignedContainer(t, lines, "[alt 成功]", map[string]int{
		"顧客": columns["顧客"], "服务": columns["服务"], "監査": columns["監査"], "数据库": columns["数据库"],
	}, useASCII, true, "fragment")
	assertAlignedContainer(t, lines, "日本語サービス", map[string]int{
		"顧客": columns["顧客"], "服务": columns["服务"],
	}, useASCII, false, "box")

	activationRow := findTextRow(lines, "更新", 0)
	active := '┃'
	if useASCII {
		active = '#'
	}
	if !isRuneCell(cellAt(lines[activationRow], columns["服务"]), active) {
		t.Errorf("useASCII=%t activation lifeline for 更新 is not aligned at 服务 column %d", useASCII, columns["服务"])
	}
	assertMessageEndpoint(t, lines, "終了", columns["新規"], useASCII, '×', 'x')
}

func participantColumns(t *testing.T, lines, labels []string, useASCII bool) map[string]int {
	t.Helper()
	headerRow := -1
	for i, line := range lines {
		allPresent := true
		for _, label := range labels {
			if !strings.Contains(line, label) {
				allPresent = false
				break
			}
		}
		if allPresent {
			headerRow = i
			break
		}
	}
	if headerRow < 0 || headerRow+1 >= len(lines) {
		t.Fatalf("participant header row not found in rendered output")
	}

	header := lines[headerRow]
	bottom := textCells(lines[headerRow+1])
	junction := '┬'
	if useASCII {
		junction = '+'
	}
	columns := make(map[string]int, len(labels))
	for _, label := range labels {
		labelColumn, ok := textColumn(header, label)
		if !ok {
			t.Fatalf("label %q has no display-cell column", label)
		}
		labelCenter := labelColumn + displayWidth(label)/2
		best, bestDistance := -1, len(bottom)+1
		for column, cell := range bottom {
			if isRuneCell(cell, junction) && abs(column-labelCenter) < bestDistance {
				best, bestDistance = column, abs(column-labelCenter)
			}
		}
		if best < 0 {
			t.Fatalf("participant %q has no header junction near display column %d", label, labelColumn)
		}
		columns[label] = best
	}
	return columns
}

func assertAlignedMessage(t *testing.T, lines []string, label string, from, to int, useASCII bool) {
	t.Helper()
	labelRow := findTextRow(lines, label, 0)
	arrowRow := findMessageRow(lines, labelRow+1, from, to, useASCII)
	if arrowRow < 0 {
		t.Errorf("message %q has no arrow row spanning display columns %d and %d", label, from, to)
		return
	}
	for _, column := range []int{from, to} {
		if isBlankCell(cellAt(lines[arrowRow], column)) {
			t.Errorf("message %q arrow is blank at display column %d", label, column)
		}
	}
}

func assertAlignedSelfMessage(t *testing.T, lines []string, label string, from int) {
	t.Helper()
	labelRow := findTextRow(lines, label, 0)
	right := from + defaultSelfMessageWidth - 1
	for row := labelRow + 1; row <= labelRow+3 && row < len(lines); row++ {
		if !isBlankCell(cellAt(lines[row], from)) && !isBlankCell(cellAt(lines[row], right)) {
			return
		}
	}
	t.Errorf("self-message %q does not keep its loop anchored at display columns %d and %d", label, from, right)
}

func assertAlignedNote(t *testing.T, lines []string, label string, first, last int, useASCII bool) {
	t.Helper()
	row := findTextRow(lines, label, 0)
	cells := textCells(lines[row])
	border := '│'
	if useASCII {
		border = '|'
	}
	labelColumn := textColumnMust(lines[row], label)
	left, right := nearestRune(cells, labelColumn, border, -1), nearestRune(cells, labelColumn, border, 1)
	if left < 0 || right < 0 || !(left < first && last < right) {
		t.Errorf("note %q does not span participant columns %d..%d: borders %d..%d", label, first, last, left, right)
		return
	}
	for _, borderRow := range []int{row - 1, row + 1} {
		if borderRow < 0 || isBlankCell(cellAt(lines[borderRow], left)) || isBlankCell(cellAt(lines[borderRow], right)) {
			t.Errorf("note %q border is not aligned with columns %d and %d", label, left, right)
		}
	}
}

func assertAlignedContainer(t *testing.T, lines []string, label string, columns map[string]int, useASCII, useFurthestRight bool, kind string) {
	t.Helper()
	row := findTextRow(lines, label, 0)
	labelColumn := textColumnMust(lines[row], label)
	cells := textCells(lines[row])
	leftRune, rightRune := '┌', '┐'
	if useASCII {
		leftRune, rightRune = '+', '+'
	}
	left := nearestRune(cells, labelColumn, leftRune, -1)
	right := nearestRune(cells, labelColumn, rightRune, 1)
	if useFurthestRight {
		right = furthestRune(cells, labelColumn, rightRune, 1)
	}
	for participant, column := range columns {
		if left < 0 || right < 0 || !(left < column && column < right) {
			t.Errorf("%s %q does not contain %s lifeline column %d: %s %d..%d", kind, label, participant, column, kind, left, right)
		}
	}
}

func assertMessageEndpoint(t *testing.T, lines []string, label string, target int, useASCII bool, unicodeRune, asciiRune rune) {
	t.Helper()
	labelRow := findTextRow(lines, label, 0)
	for row := labelRow + 1; row < len(lines) && row <= labelRow+8; row++ {
		if !isBlankCell(cellAt(lines[row], target)) && (hasCellContent(cellAt(lines[row], target), unicodeRune) || hasCellContent(cellAt(lines[row], target), asciiRune)) {
			return
		}
	}
	t.Errorf("message %q has no expected endpoint at display column %d (ASCII=%t)", label, target, useASCII)
}

func findTextRow(lines []string, text string, start int) int {
	for i := start; i < len(lines); i++ {
		if strings.Contains(lines[i], text) {
			return i
		}
	}
	return -1
}

func findMessageRow(lines []string, start, from, to int, useASCII bool) int {
	for row := start; row < len(lines) && row <= start+8; row++ {
		cells := textCells(lines[row])
		lo, hi := from, to
		if lo > hi {
			lo, hi = hi, lo
		}
		strokes := 0
		for column := lo; column <= hi; column++ {
			if isHorizontalCell(cellAtCells(cells, column), useASCII) {
				strokes++
			}
		}
		if strokes > 0 && !isBlankCell(cellAtCells(cells, from)) && !isBlankCell(cellAtCells(cells, to)) {
			return row
		}
	}
	return -1
}

func textColumnMust(line, text string) int {
	column, ok := textColumn(line, text)
	if !ok {
		return -1
	}
	return column
}

func textColumn(line, text string) (int, bool) {
	cells := textCells(line)
	for start, cell := range cells {
		if !isTextualCell(cell) {
			continue
		}
		var got strings.Builder
		for end := start; end < len(cells); end++ {
			if !cells[end].continuation {
				got.WriteString(cells[end].content)
			}
			value := got.String()
			if strings.HasPrefix(value, text) {
				return start, true
			}
			if !strings.HasPrefix(text, value) {
				break
			}
		}
	}
	return 0, false
}

func nearestRune(cells []textCell, start int, target rune, direction int) int {
	for column := start + direction; column >= 0 && column < len(cells); column += direction {
		if isRuneCell(cells[column], target) {
			return column
		}
	}
	return -1
}

func furthestRune(cells []textCell, start int, target rune, direction int) int {
	found := -1
	for column := start + direction; column >= 0 && column < len(cells); column += direction {
		if isRuneCell(cells[column], target) {
			found = column
		}
	}
	return found
}

func cellAt(line string, column int) textCell { return cellAtCells(textCells(line), column) }

func cellAtCells(cells []textCell, column int) textCell {
	if column < 0 || column >= len(cells) {
		return cellRune(' ')
	}
	return cells[column]
}

func isBlankCell(cell textCell) bool {
	return !cell.textual && !cell.continuation && cell.content == " "
}

func hasCellContent(cell textCell, content rune) bool {
	return !cell.continuation && cell.content == string(content)
}

func isHorizontalCell(cell textCell, useASCII bool) bool {
	if cell.textual || cell.continuation {
		return false
	}
	if useASCII {
		return cell.content == "-" || cell.content == "."
	}
	return cell.content == "─" || cell.content == "┈"
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func TestPutTextLeadingCombiningMarkSkipsPadding(t *testing.T) {
	line := []textCell{cellRune(' '), cellRune(' '), cellRune(' '), cellRune('|')}
	putText(line, 2, "\u0301A")

	if line[1] != cellRune(' ') {
		t.Fatalf("leading mark attached to padding cell: %#v", line[1])
	}
	if line[2] != textualCell("\u0301A") {
		t.Fatalf("text cell = %#v, want %#v", line[2], textualCell("\u0301A"))
	}

	line = []textCell{cellRune(' '), cellRune(' '), cellRune(' ')}
	putText(line, 1, "\u0301")
	if line[1] != cellRune(' ') {
		t.Fatalf("leading mark mutated padding: %#v", line[1])
	}
}

func TestPutTextKeepsLeadingMarksWithBase(t *testing.T) {
	for _, tc := range []struct {
		name string
		text string
		want []textCell
	}{
		{name: "ASCII base", text: "\u0301A", want: []textCell{cellRune(' '), textualCell("\u0301A")}},
		{name: "wide base", text: "\u0301用́", want: []textCell{cellRune(' '), textualCell("\u0301用́"), continuationCell}},
		{name: "format mark", text: "\u200bA", want: []textCell{cellRune(' '), textualCell("\u200bA")}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			line := make([]textCell, 8)
			for i := range line {
				line[i] = cellRune(' ')
			}
			putText(line, 1, tc.text)
			for i, want := range tc.want {
				if line[i] != want {
					t.Errorf("cell[%d] = %#v, want %#v", i, line[i], want)
				}
			}
		})
	}
}

func TestLiteralControlCharacterPreservedInRendererText(t *testing.T) {
	control := "\x01"
	if got := trimCells(textCells("left" + control + "right")); got != "left"+control+"right" {
		t.Fatalf("literal control character lost from cell text: %q", got)
	}

	a := &Participant{ID: "a", Label: "A" + control + "ctor", Index: 0}
	b := &Participant{ID: "b", Label: "B", Index: 1}
	sd := &SequenceDiagram{
		Participants: []*Participant{a, b},
		Boxes:        []*Box{{Title: "box" + control + "title", First: 0, Last: 1}},
		Events: []Event{
			{Kind: EventFragmentStart, Fragment: &Fragment{Type: FragmentLoop, Label: "fragment" + control + "label"}},
			{Kind: EventMessage, Message: &Message{From: a, To: b, Label: "send" + control + "message", ArrowType: SolidArrow}},
			{Kind: EventFragmentEnd},
			{Kind: EventNote, Note: &Note{Placement: NoteOver, Participants: []*Participant{a}, Text: "note" + control + "text"}},
		},
	}
	for _, useASCII := range []bool{false, true} {
		config := diagram.NewTestConfig(useASCII, "cli")
		got, err := Render(sd, config)
		if err != nil {
			t.Fatalf("Render(useASCII=%t): %v", useASCII, err)
		}
		for _, want := range []string{"A" + control + "ctor", "send" + control + "message", "note" + control + "text", "box" + control + "title", "fragment" + control + "label"} {
			if !strings.Contains(got, want) {
				t.Errorf("Render(useASCII=%t) missing literal control text %q in %q", useASCII, want, got)
			}
		}
	}
}

func TestTextCellsAttachCombiningMarks(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input string
		cells []textCell
	}{
		{name: "ASCII base", input: "é", cells: []textCell{textualCell("é")}},
		{name: "leading combining mark", input: "\u0301A", cells: []textCell{textualCell("\u0301A")}},
		{name: "mark after padding", input: "  \u0301A", cells: []textCell{cellRune(' '), cellRune(' '), textualCell("\u0301A")}},
		{name: "zero-width format", input: "A\u200bB", cells: []textCell{textualCell("A\u200b"), textualCell("B")}},
		{name: "mark after border", input: "| \u0301A", cells: []textCell{cellRune('|'), cellRune(' '), textualCell("\u0301A")}},
		{name: "leading mark before border", input: "\u0301| A", cells: []textCell{cellRune('|'), cellRune(' '), textualCell("\u0301A")}},
		{name: "CJK base", input: "用́", cells: []textCell{textualCell("用́"), continuationCell}},
		{name: "mixed text", input: "A用́B", cells: []textCell{textualCell("A"), textualCell("用́"), continuationCell, textualCell("B")}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := textCells(tc.input)
			if len(got) != len(tc.cells) {
				t.Fatalf("textCells(%q) has %d cells, want %d: %#v", tc.input, len(got), len(tc.cells), got)
			}
			for i := range tc.cells {
				if got[i] != tc.cells[i] {
					t.Errorf("textCells(%q)[%d] = %#v, want %#v", tc.input, i, got[i], tc.cells[i])
				}
			}
			if displayWidth(tc.input) != len(got) {
				t.Errorf("displayWidth(%q) = %d, want cell count %d", tc.input, displayWidth(tc.input), len(got))
			}
		})
	}
}

// verifySequenceDiagramWithCharset verifies a test case with the specified charset.
func verifySequenceDiagramWithCharset(t *testing.T, testCaseFile string, useAscii bool) {
	tc, err := testutil.ReadSequenceTestCase(testCaseFile)
	if err != nil {
		t.Fatalf("Failed to read test case file: %v", err)
	}

	sd, err := Parse(tc.Mermaid)
	if err != nil {
		t.Fatalf("Failed to parse sequence diagram: %v", err)
	}

	// Create config with specified charset
	config := diagram.NewTestConfig(useAscii, "cli")

	actual, err := Render(sd, config)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	expectedNormalized := testutil.NormalizeWhitespace(tc.Expected)
	actualNormalized := testutil.NormalizeWhitespace(actual)

	if expectedNormalized != actualNormalized {
		expectedWithSpaces := testutil.VisualizeWhitespace(expectedNormalized)
		actualWithSpaces := testutil.VisualizeWhitespace(actualNormalized)
		t.Errorf("Sequence diagram didn't match\nExpected:\n%v\nActual:\n%v", expectedWithSpaces, actualWithSpaces)
	}
}
