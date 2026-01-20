package cmd

import (
	"strings"
	"testing"

	"github.com/AlexanderGrooff/mermaid-ascii/internal/diagram"
)

func TestGraphFitRespectsMaxWidth(t *testing.T) {
	input := "flowchart TB\n" +
		"A[Very Long Node Label Here] --> B[Another Long Label]\n"
	cfg := diagram.NewTestConfig(true, "cli")
	cfg.MaxWidth = 20
	cfg.FitPolicy = diagram.FitPolicyAuto

	out, err := RenderDiagram(input, cfg)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if maxOutputLineWidth(out) > cfg.MaxWidth {
		t.Fatalf("expected width <= %d, got %d", cfg.MaxWidth, maxOutputLineWidth(out))
	}
	if !strings.Contains(out, "Very") {
		t.Fatalf("expected label content retained")
	}
}
