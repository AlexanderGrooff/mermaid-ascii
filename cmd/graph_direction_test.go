package cmd

import (
	"strings"
	"testing"

	"github.com/AlexanderGrooff/mermaid-ascii/internal/diagram"
)

func TestGraphDirectionOverride(t *testing.T) {
	input := "graph LR\nA --> B\n"
	cfg := diagram.NewTestConfig(true, "cli")
	cfg.GraphDirection = "TD"
	out, err := RenderDiagram(input, cfg)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if !containsVerticalFlow(out) {
		t.Fatalf("expected TD render to be vertical")
	}
}

func containsVerticalFlow(out string) bool {
	lines := strings.Split(out, "\n")
	idxA := -1
	idxB := -1
	for i, line := range lines {
		if strings.Contains(line, "| A |") {
			idxA = i
		}
		if strings.Contains(line, "| B |") {
			idxB = i
		}
		if strings.Contains(line, "| A |") && strings.Contains(line, "| B |") {
			return false
		}
	}
	return idxA != -1 && idxB != -1 && idxB > idxA
}
