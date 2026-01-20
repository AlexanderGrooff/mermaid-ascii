package cmd

import (
	"strings"
	"testing"

	"github.com/AlexanderGrooff/mermaid-ascii/internal/diagram"
)

func TestGraphEdgeLabelEllipsis(t *testing.T) {
	input := "graph LR\nA -->|This is a long edge label| B\n"
	cfg := diagram.NewTestConfig(true, "cli")
	cfg.EdgeLabelPolicy = "ellipsis"
	cfg.EdgeLabelMaxWidth = 6
	out, err := RenderDiagram(input, cfg)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if !strings.Contains(out, "...") {
		t.Fatalf("expected ellipsis edge label")
	}
	if strings.Contains(out, "This is a long edge label") {
		t.Fatalf("expected edge label to be truncated")
	}
}
