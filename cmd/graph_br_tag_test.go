package cmd

import (
"strings"
"testing"

"github.com/AlexanderGrooff/mermaid-ascii/internal/diagram"
)

// TestBRTagConversion tests that <br/> and <br> tags are converted to newlines
func TestBRTagConversion(t *testing.T) {
tests := []struct {
name     string
input    string
expected []string // Expected strings to appear on separate lines
}{
{
name:     "br_with_slash",
input:    "graph LR\nA[\"Line 1<br/>Line 2\"]\n",
expected: []string{"Line 1", "Line 2"},
},
{
name:     "br_without_slash",
input:    "graph LR\nA[\"Line 1<br>Line 2\"]\n",
expected: []string{"Line 1", "Line 2"},
},
{
name:     "multiple_br_tags",
input:    "graph LR\nA[\"First<br/>Second<br/>Third\"]\n",
expected: []string{"First", "Second", "Third"},
},
{
name:     "mixed_br_tags",
input:    "graph LR\nA[\"Line 1<br>Line 2<br/>Line 3\"]\n",
expected: []string{"Line 1", "Line 2", "Line 3"},
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
cfg := diagram.NewTestConfig(true, "cli")
cfg.LabelWrapWidth = 0 // Disable automatic wrapping
out, err := RenderDiagram(tt.input, cfg)
if err != nil {
t.Fatalf("render failed: %v", err)
}

// Check that all expected strings appear on different lines
for i := 0; i < len(tt.expected)-1; i++ {
if !appearsOnDifferentLines(out, tt.expected[i], tt.expected[i+1]) {
t.Errorf("expected %q and %q to appear on different lines\nOutput:\n%s",
tt.expected[i], tt.expected[i+1], out)
}
}
})
}
}

// TestBRTagWithWrapping tests that <br/> tags work in combination with label wrapping
func TestBRTagWithWrapping(t *testing.T) {
input := "graph LR\nA[\"First Line<br/>Second Line with long text that should wrap\"]\n"
cfg := diagram.NewTestConfig(true, "cli")
cfg.LabelWrapWidth = 15 // Enable wrapping at 15 chars
out, err := RenderDiagram(input, cfg)
if err != nil {
t.Fatalf("render failed: %v", err)
}

// Check that "First Line" and "Second Line" appear on different lines
if !appearsOnDifferentLines(out, "First Line", "Second Line") {
t.Errorf("expected <br/> to create separate lines even with wrapping enabled\nOutput:\n%s", out)
}
}

// TestBRTagInMultipleNodes tests <br/> tags in multiple nodes
func TestBRTagInMultipleNodes(t *testing.T) {
input := `graph LR
A["Node A<br/>Line 2"]
B["Node B<br/>Line 2"]
A --> B
`
cfg := diagram.NewTestConfig(true, "cli")
cfg.LabelWrapWidth = 0
out, err := RenderDiagram(input, cfg)
if err != nil {
t.Fatalf("render failed: %v", err)
}

// Check that both nodes have multi-line labels
if !appearsOnDifferentLines(out, "Node A", "Line 2") {
t.Errorf("expected Node A to have multi-line label\nOutput:\n%s", out)
}
if !appearsOnDifferentLines(out, "Node B", "Line 2") {
t.Errorf("expected Node B to have multi-line label\nOutput:\n%s", out)
}
}

// TestLiteralNewline tests that literal \n still works
func TestLiteralNewline(t *testing.T) {
input := "graph LR\nA[\"Line 1\\nLine 2\"]\n"
cfg := diagram.NewTestConfig(true, "cli")
cfg.LabelWrapWidth = 0
out, err := RenderDiagram(input, cfg)
if err != nil {
t.Fatalf("render failed: %v", err)
}

if !appearsOnDifferentLines(out, "Line 1", "Line 2") {
t.Errorf("expected literal \\n to create separate lines\nOutput:\n%s", out)
}
}

// appearsOnDifferentLines checks if two strings appear on different lines in the output
func appearsOnDifferentLines(output, first, second string) bool {
lines := strings.Split(output, "\n")
firstLine := -1
secondLine := -1

for i, line := range lines {
if strings.Contains(line, first) && firstLine == -1 {
firstLine = i
}
if strings.Contains(line, second) && secondLine == -1 {
secondLine = i
}
}

// Both strings found and on different lines
return firstLine != -1 && secondLine != -1 && firstLine != secondLine
}
