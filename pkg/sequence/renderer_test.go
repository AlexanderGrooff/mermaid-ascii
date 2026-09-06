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
	}
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
