package graph

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlexanderGrooff/mermaid-ascii/pkg/diagram"
	"github.com/AlexanderGrooff/mermaid-ascii/pkg/diagram/testutil"
	"github.com/mattn/go-runewidth"
	log "github.com/sirupsen/logrus"
)

func verifyMap(t *testing.T, testCaseFile string, useAscii bool) {
	tc, err := testutil.ReadTestCase(testCaseFile)
	if err != nil {
		t.Fatalf("Failed to read test case file: %v", err)
	}

	properties, err := Parse(tc.Mermaid, "cli")
	if err != nil {
		log.Fatal("Failed to parse mermaid: ", err)
	}
	properties.paddingX = tc.PaddingX
	properties.paddingY = tc.PaddingY
	properties.useAscii = useAscii
	actualMap := Draw(properties)
	if tc.MaxWidth > 0 && maxDisplayWidth(actualMap) > tc.MaxWidth {
		properties.paddingX = 1
		properties.paddingY = 1
		actualMap = Draw(properties)
	}
	if tc.Expected != actualMap {
		expectedWithSpaces := testutil.VisualizeWhitespace(tc.Expected)
		actualWithSpaces := testutil.VisualizeWhitespace(actualMap)
		t.Errorf("Map didn't match actual map\nExpected:\n%v\nActual:\n%v", expectedWithSpaces, actualWithSpaces)
	}
}

func maxDisplayWidth(output string) int {
	width := 0
	for _, line := range strings.Split(output, "\n") {
		width = max(width, runewidth.StringWidth(line))
	}
	return width
}

func TestASCII(t *testing.T) {
	dir := "testdata/ascii"
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("Failed to read directory %s: %v", dir, err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".txt") {
			t.Run(file.Name(), func(t *testing.T) {
				verifyMap(t, filepath.Join(dir, file.Name()), true)
			})
		}
	}
}

func TestExtendedChars(t *testing.T) {
	dir := "testdata/extended-chars"
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("Failed to read directory %s: %v", dir, err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".txt") {
			t.Run(file.Name(), func(t *testing.T) {
				verifyMap(t, filepath.Join(dir, file.Name()), false)
			})
		}
	}
}

// TestMultibyte verifies that multibyte UTF-8 labels (Cyrillic, Greek, accented
// Latin, etc.) render correctly without splitting runes into invalid byte
// fragments. Test cases are loaded from testdata/multibyte/*.txt.
func TestMultibyte(t *testing.T) {
	dir := "testdata/multibyte"
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("Failed to read directory %s: %v", dir, err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".txt") {
			t.Run(file.Name(), func(t *testing.T) {
				verifyMap(t, filepath.Join(dir, file.Name()), true)
			})
		}
	}
}

// TestGraphUseAsciiConfig tests that RenderDiagram respects config.UseAscii for graphs
func TestGraphUseAsciiConfig(t *testing.T) {
	mermaidInput := `graph LR
A --> B`

	// Test with UseAscii = true (should produce ASCII output)
	asciiConfig := &diagram.Config{
		UseAscii:         true,
		BoxBorderPadding: 1,
		PaddingBetweenX:  5,
		PaddingBetweenY:  5,
		GraphDirection:   "LR",
		StyleType:        "cli",
	}
	asciiOutput, err := renderGraph(mermaidInput, asciiConfig)
	if err != nil {
		t.Fatalf("Failed to render with ASCII config: %v", err)
	}

	// Test with UseAscii = false (should produce Unicode output)
	unicodeConfig := &diagram.Config{
		UseAscii:         false,
		BoxBorderPadding: 1,
		PaddingBetweenX:  5,
		PaddingBetweenY:  5,
		GraphDirection:   "LR",
		StyleType:        "cli",
	}
	unicodeOutput, err := renderGraph(mermaidInput, unicodeConfig)
	if err != nil {
		t.Fatalf("Failed to render with Unicode config: %v", err)
	}

	// Verify outputs are different
	if asciiOutput == unicodeOutput {
		t.Errorf("ASCII and Unicode outputs should be different, but they were identical:\n%s", asciiOutput)
	}

	// Verify ASCII output contains ASCII characters (not Unicode box-drawing)
	if strings.Contains(asciiOutput, "┌") || strings.Contains(asciiOutput, "─") || strings.Contains(asciiOutput, "│") {
		t.Errorf("ASCII output should not contain Unicode box-drawing characters, but found:\n%s", asciiOutput)
	}

	// Verify Unicode output contains Unicode characters
	if !strings.Contains(unicodeOutput, "┌") && !strings.Contains(unicodeOutput, "─") && !strings.Contains(unicodeOutput, "│") {
		t.Errorf("Unicode output should contain Unicode box-drawing characters, but found:\n%s", unicodeOutput)
	}
}

// Sequence diagram tests moved to sequence_test.go
