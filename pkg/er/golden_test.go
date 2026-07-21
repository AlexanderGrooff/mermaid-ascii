package er

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/AlexanderGrooff/mermaid-ascii/pkg/diagram/testutil"
)

// erTestDataPath returns the absolute path to a cmd/testdata subdirectory,
// resolved from this file's location so tests work from any working directory.
func erTestDataPath(subdir string) string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "cmd", "testdata", subdir)
}

// verifyErGolden renders the mermaid input from a golden file and compares it
// against the expected output stored in the same file.
func verifyErGolden(t *testing.T, testCaseFile string, useAscii bool) {
	tc, err := testutil.ReadSequenceTestCase(testCaseFile)
	if err != nil {
		t.Fatalf("Failed to read test case file: %v", err)
	}

	d, err := Parse(tc.Mermaid)
	if err != nil {
		t.Fatalf("Failed to parse er diagram: %v", err)
	}

	actual := Render(d, useAscii)

	expectedNormalized := testutil.NormalizeWhitespace(tc.Expected)
	actualNormalized := testutil.NormalizeWhitespace(actual)

	if expectedNormalized != actualNormalized {
		expectedWithSpaces := testutil.VisualizeWhitespace(expectedNormalized)
		actualWithSpaces := testutil.VisualizeWhitespace(actualNormalized)
		t.Errorf("Er diagram didn't match\nExpected:\n%v\nActual:\n%v", expectedWithSpaces, actualWithSpaces)
	}
}

// runErGoldenDir runs every .txt golden file in a testdata subdirectory.
// New cases are picked up automatically — just drop a file in the directory.
func runErGoldenDir(t *testing.T, subdir string, useAscii bool) {
	dir := erTestDataPath(subdir)
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("Failed to read directory %s: %v", dir, err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".txt") {
			t.Run(file.Name(), func(t *testing.T) {
				verifyErGolden(t, filepath.Join(dir, file.Name()), useAscii)
			})
		}
	}
}

// TestErDiagramRendering tests all er diagram golden files with Unicode charset.
func TestErDiagramRendering(t *testing.T) {
	runErGoldenDir(t, "er", false)
}

// TestErDiagramRendering_ASCII tests er diagram golden files with ASCII charset.
func TestErDiagramRendering_ASCII(t *testing.T) {
	runErGoldenDir(t, "er-ascii", true)
}

// TestErDiagramRendering_ASCIISmokeTest verifies ASCII rendering works for every
// Unicode test input, without requiring an ASCII golden file for each case.
func TestErDiagramRendering_ASCIISmokeTest(t *testing.T) {
	dir := erTestDataPath("er")
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("Failed to read directory %s: %v", dir, err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".txt") {
			t.Run(file.Name(), func(t *testing.T) {
				tc, err := testutil.ReadSequenceTestCase(filepath.Join(dir, file.Name()))
				if err != nil {
					t.Fatalf("Failed to read test case file: %v", err)
				}
				d, err := Parse(tc.Mermaid)
				if err != nil {
					t.Fatalf("Failed to parse er diagram: %v", err)
				}
				if out := Render(d, true); strings.TrimSpace(out) == "" {
					t.Errorf("ASCII render produced empty output")
				}
			})
		}
	}
}
