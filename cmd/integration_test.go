package cmd

import (
	"github.com/AlexanderGrooff/mermaid-ascii/pkg/sequence"
	"strings"
	"testing"

	"github.com/AlexanderGrooff/mermaid-ascii/pkg/diagram"
)

// TestSequenceDiagramIntegration tests end-to-end rendering of sequence diagrams.
func TestSequenceDiagramIntegration(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantSubstring []string // Substrings that must appear in output
		wantNoError   bool
	}{
		{
			name: "simple two participant diagram",
			input: `sequenceDiagram
    Alice->>Bob: Hello
    Bob-->>Alice: Hi`,
			wantSubstring: []string{"Alice", "Bob", "Hello", "Hi", "│"},
			wantNoError:   true,
		},
		{
			name: "self-message diagram",
			input: `sequenceDiagram
    Alice->>Alice: Think`,
			wantSubstring: []string{"Alice", "Think"},
			wantNoError:   true,
		},
		{
			name: "three participants",
			input: `sequenceDiagram
    Alice->>Bob: Request
    Bob->>Charlie: Forward
    Charlie-->>Bob: Response
    Bob-->>Alice: Done`,
			wantSubstring: []string{"Alice", "Bob", "Charlie", "Request", "Forward", "Response", "Done"},
			wantNoError:   true,
		},
		{
			name: "with explicit participants",
			input: `sequenceDiagram
    participant A as Alice
    participant B as Bob
    A->>B: Test`,
			wantSubstring: []string{"Alice", "Bob", "Test"},
			wantNoError:   true,
		},
		{
			name: "with comments",
			input: `sequenceDiagram
    %% This is a comment
    Alice->>Bob: Hello
    %% Another comment`,
			wantSubstring: []string{"Alice", "Bob", "Hello"},
			wantNoError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := diagram.NewTestConfig(false, "cli") // Unicode, CLI style

			output, err := RenderDiagram(tt.input, config)

			if tt.wantNoError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !tt.wantNoError && err == nil {
				t.Error("Expected error but got none")
				return
			}

			// Check all expected substrings are present
			for _, want := range tt.wantSubstring {
				if !strings.Contains(output, want) {
					t.Errorf("Output missing expected substring %q\nOutput:\n%s", want, output)
				}
			}

			// Verify output is not empty
			if tt.wantNoError && len(output) == 0 {
				t.Error("Output is empty but expected content")
			}
		})
	}
}

// TestSequenceDiagramIntegration_ASCII tests ASCII rendering mode.
func TestSequenceDiagramIntegration_ASCII(t *testing.T) {
	input := `sequenceDiagram
    Alice->>Bob: Hello`

	config := diagram.NewTestConfig(true, "cli") // ASCII, CLI style

	output, err := RenderDiagram(input, config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// ASCII mode should use + and | characters
	if !strings.Contains(output, "+") {
		t.Error("ASCII output should contain '+' characters")
	}
	if !strings.Contains(output, "|") {
		t.Error("ASCII output should contain '|' characters")
	}
	// Should not contain Unicode box-drawing characters
	if strings.Contains(output, "│") || strings.Contains(output, "┌") {
		t.Error("ASCII output should not contain Unicode box-drawing characters")
	}
}

// TestSequenceDiagramIntegration_ErrorHandling tests error cases.
func TestSequenceDiagramIntegration_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldError bool
	}{
		{
			name:        "empty input - defaults to graph, fails graph parsing",
			input:       "",
			shouldError: true,
		},
		{
			name:        "invalid syntax - defaults to graph, fails graph parsing",
			input:       "not a diagram",
			shouldError: true,
		},
		{
			name:        "sequence diagram with only comments",
			input:       "sequenceDiagram\n%% Only comments",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := diagram.NewTestConfig(false, "cli") // Unicode, CLI style
			_, err := RenderDiagram(tt.input, config)

			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestDiagramFactoryIntegration tests diagram type detection.
func TestDiagramFactoryIntegration(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedType string
	}{
		{
			name: "sequence diagram",
			input: `sequenceDiagram
    A->>B: Test`,
			expectedType: "sequence",
		},
		{
			name: "graph diagram",
			input: `graph LR
    A-->B`,
			expectedType: "graph",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diag, err := DiagramFactory(tt.input)
			if err != nil {
				t.Fatalf("Factory error: %v", err)
			}

			if diag.Type() != tt.expectedType {
				t.Errorf("Expected type %q, got %q", tt.expectedType, diag.Type())
			}

			// Verify it can parse and render
			if err := diag.Parse(tt.input); err != nil {
				t.Errorf("Parse error: %v", err)
			}

			config := diagram.NewTestConfig(false, "cli")
			output, err := diag.Render(config)
			if err != nil {
				t.Errorf("Render error: %v", err)
			}

			if len(output) == 0 {
				t.Error("Render produced empty output")
			}
		})
	}
}

// BenchmarkSequenceDiagramRendering benchmarks the rendering performance.
func BenchmarkSequenceDiagramRendering(b *testing.B) {
	input := `sequenceDiagram
    participant Alice
    participant Bob
    participant Charlie
    Alice->>Bob: Message 1
    Bob->>Charlie: Message 2
    Charlie-->>Bob: Response 1
    Bob-->>Alice: Response 2`

	config := diagram.NewTestConfig(false, "cli")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := RenderDiagram(input, config)
		if err != nil {
			b.Fatalf("Render error: %v", err)
		}
	}
}

// BenchmarkSequenceDiagramParsing benchmarks just the parsing performance.
func BenchmarkSequenceDiagramParsing(b *testing.B) {
	input := `sequenceDiagram
    participant Alice
    participant Bob
    Alice->>Bob: Test
    Bob-->>Alice: Response`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := sequence.Parse(input)
		if err != nil {
			b.Fatalf("Parse error: %v", err)
		}
	}
}

// TestRenderConfig_Defaults tests that default configuration works correctly.
func TestRenderConfig_Defaults(t *testing.T) {
	config := diagram.DefaultConfig()

	// Verify sensible defaults
	if config.UseAscii {
		t.Error("Default should be Unicode, not ASCII")
	}
	if config.SequenceParticipantSpacing <= 0 {
		t.Error("SequenceParticipantSpacing should have positive default")
	}
	if config.SequenceMessageSpacing < 0 {
		t.Error("SequenceMessageSpacing should be non-negative")
	}

	// Test that defaults actually work for rendering
	input := `sequenceDiagram
    A->>B: Test`

	testConfig := diagram.NewTestConfig(false, "cli")
	output, err := RenderDiagram(input, testConfig)
	if err != nil {
		t.Errorf("Default config failed to render: %v", err)
	}
	if len(output) == 0 {
		t.Error("Default config produced empty output")
	}
}
