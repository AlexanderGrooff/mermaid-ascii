package sequence

import (
	"testing"

	"github.com/AlexanderGrooff/mermaid-ascii/internal/diagram"
)

func TestSequenceFitRespectsMaxWidth(t *testing.T) {
	input := "sequenceDiagram\nAlice->>Bob: This is a very long message label\n"
	cfg := diagram.DefaultConfig()
	cfg.UseAscii = true
	cfg.MaxWidth = 40
	cfg.FitPolicy = diagram.FitPolicyAuto

	parsed, err := Parse(input)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	out, err := Render(parsed, cfg)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if maxOutputLineWidth(out) > cfg.MaxWidth {
		t.Fatalf("expected width <= %d, got %d", cfg.MaxWidth, maxOutputLineWidth(out))
	}
}
