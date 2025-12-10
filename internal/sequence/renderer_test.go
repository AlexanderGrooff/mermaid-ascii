package sequence

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/AlexanderGrooff/mermaid-ascii/internal/diagram"
	"github.com/AlexanderGrooff/mermaid-ascii/internal/diagram/testutil"
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
		"bidirectional_messages.txt",
		"dotted_arrows_only.txt",
		"four_participants.txt",
		"long_participant_names.txt",
		"messages_without_labels.txt",
		"multiword_labels.txt",
		"self_message.txt",
		"simple_two_participants.txt",
		"single_message.txt",
		"three_participants.txt",
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
		"simple_two_participants.txt",
		"dotted_arrows_only.txt",
		"self_message.txt",
		"three_participants.txt",
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
		"bidirectional_messages.txt",
		"dotted_arrows_only.txt",
		"four_participants.txt",
		"long_participant_names.txt",
		"messages_without_labels.txt",
		"multiword_labels.txt",
		"self_message.txt",
		"simple_two_participants.txt",
		"single_message.txt",
		"three_participants.txt",
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
