package cmd

import "testing"

func TestRenderDiagramWithOptions(t *testing.T) {
	input := "graph LR\nA[This is a long label] --> B[Another long label]\n"
	out, err := RenderDiagramWithOptions(input, WithMaxWidth(10), WithAscii())
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if maxOutputLineWidth(out) > 10 {
		t.Fatalf("expected width <= 10, got %d", maxOutputLineWidth(out))
	}
}
