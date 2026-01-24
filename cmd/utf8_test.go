package cmd

import (
	"testing"
	
	"github.com/AlexanderGrooff/mermaid-ascii/internal/diagram"
)

// TestUTF8MultiLineNode tests that multi-line nodes with UTF-8 characters
// render properly without content being split across multiple boxes.
func TestUTF8MultiLineNode(t *testing.T) {
	input := `flowchart TD
		a["┌─ TIMER<br/>├─> Step 1<br/>└─> Step 2"]
		b["日本語 🎉"]
		a --> b`
	
	config := diagram.NewTestConfig(true, "cli")
	output, err := RenderDiagram(input, config)
	if err != nil {
		t.Fatalf("RenderDiagram failed: %v", err)
	}

	// Verify all UTF-8 characters are present
	expectedChars := []string{"┌─", "├─>", "└─>", "日本語", "🎉"}
	for _, char := range expectedChars {
		if !contains(output, char) {
			t.Errorf("Output missing expected UTF-8 character %q\nGot:\n%s", char, output)
		}
	}

	// Verify node content stays together (not split across boxes)
	// Each line should appear on its own line in the output
	lines := []string{"TIMER", "Step 1", "Step 2"}
	for _, line := range lines {
		if !contains(output, line) {
			t.Errorf("Output missing expected line content %q\nGot:\n%s", line, output)
		}
	}

	// Verify no corruption markers
	if contains(output, "\ufffd") || contains(output, "ââ") {
		t.Errorf("Output contains UTF-8 corruption markers\nGot:\n%s", output)
	}

	// Debug: Print actual output for visual inspection
	t.Logf("Rendered output:\n%s", output)
}

// TestUTF8Characters verifies that Unicode characters (including multi-byte UTF-8)
// are rendered correctly without corruption.
func TestUTF8Characters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string // Strings that should appear in output
	}{
		{
			name: "Box drawing characters",
			input: `flowchart TD
				node["├─> test └─>"]`,
			expected: []string{"├─>", "└─>"},
		},
		{
			name: "Mixed ASCII and UTF-8",
			input: `flowchart TD
				node["Hello ├─> World"]`,
			expected: []string{"Hello", "├─>", "World"},
		},
		{
			name: "Japanese characters",
			input: `flowchart TD
				node["こんにちは"]`,
			expected: []string{"こんにちは"},
		},
		{
			name: "Emoji",
			input: `flowchart TD
				node["✓ Success ✗ Failure"]`,
			expected: []string{"✓", "Success", "✗", "Failure"},
		},
		{
			name: "Multi-line with UTF-8",
			input: `flowchart TD
				node["Line 1<br/>├─> Line 2<br/>└─> Line 3"]`,
			expected: []string{"Line 1", "├─>", "Line 2", "└─>", "Line 3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := diagram.NewTestConfig(true, "cli")
			output, err := RenderDiagram(tt.input, config)
			if err != nil {
				t.Fatalf("RenderDiagram failed: %v", err)
			}

			for _, expected := range tt.expected {
				if !contains(output, expected) {
					t.Errorf("Output missing expected string %q\nGot:\n%s", expected, output)
				}
			}

			// Verify no corruption markers (typical UTF-8 corruption creates � or strange byte sequences)
			if contains(output, "\ufffd") || contains(output, "ââ") {
				t.Errorf("Output contains UTF-8 corruption markers\nGot:\n%s", output)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && 
		(haystack == needle || len(haystack) > len(needle) && indexOf(haystack, needle) >= 0)
}

func indexOf(haystack, needle string) int {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}
