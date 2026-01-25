package cmd

import (
	"strings"
	"testing"

	"github.com/AlexanderGrooff/mermaid-ascii/internal/diagram"
)

func TestGraphLabelWrapping(t *testing.T) {
	input := "graph LR\nA[This is a long label]\nA --> B\n"
	cfg := diagram.NewTestConfig(true, "cli")
	cfg.LabelWrapWidth = 8
	out, err := RenderDiagram(input, cfg)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if !containsWrappedLabel(out, "This is", "a long") {
		t.Fatalf("expected wrapped label to span multiple lines")
	}
}

func containsWrappedLabel(out, first, second string) bool {
	lines := strings.Split(out, "\n")
	firstIdx := -1
	secondIdx := -1
	for i, line := range lines {
		if strings.Contains(line, first) {
			firstIdx = i
		}
		if strings.Contains(line, second) {
			secondIdx = i
		}
	}
	return firstIdx != -1 && secondIdx != -1 && firstIdx != secondIdx
}
