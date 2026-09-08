package render

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlexanderGrooff/mermaid-ascii/pkg/diagram"
	"github.com/AlexanderGrooff/mermaid-ascii/pkg/diagram/testutil"
	"github.com/mattn/go-runewidth"
)

func TestGraphMaxWidthGolden(t *testing.T) {
	tests := []struct {
		name      string
		compacted bool
		met       bool
	}{
		{name: "normal fit", met: true},
		{name: "compact fit", compacted: true, met: true},
		{name: "cannot fit", compacted: true, met: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join("..", "graph", "testdata", "ascii", strings.ReplaceAll(tc.name, " ", "_")+".txt")
			golden, err := testutil.ReadTestCase(path)
			if err != nil {
				t.Fatal(err)
			}

			config := diagram.NewTestConfig(true, "cli")
			config.MaxWidth = golden.MaxWidth
			output, status, err := RenderDiagramWithStatus(golden.Mermaid, config)
			if err != nil {
				t.Fatal(err)
			}
			if output != golden.Expected {
				t.Fatalf("golden output mismatch\nexpected:\n%s\nactual:\n%s", golden.Expected, output)
			}
			if status.Compacted != tc.compacted || status.Met != tc.met {
				t.Fatalf("status = %+v, want compacted=%v met=%v", status, tc.compacted, tc.met)
			}
			if got := maxDisplayWidth(output); got != status.Width {
				t.Fatalf("status width = %d, measured %d", status.Width, got)
			}
		})
	}
}

func TestMaxWidthHandlesUnicodeClusterLabels(t *testing.T) {
	for _, label := range []string{"👩‍💻", "á́́́́"} {
		t.Run(label, func(t *testing.T) {
			config := diagram.NewTestConfig(true, "cli")
			config.MaxWidth = 12
			output, status, err := RenderDiagramWithStatus("graph LR\nA[\""+label+"\"] --> B", config)
			if err != nil {
				t.Fatalf("RenderDiagramWithStatus() error = %v", err)
			}
			if !status.Compacted || !status.Met {
				t.Fatalf("status = %+v, want compact fit", status)
			}
			if status.Width > config.MaxWidth {
				t.Fatalf("width = %d, want <= %d", status.Width, config.MaxWidth)
			}
			if !strings.Contains(output, label) || !strings.Contains(output, " B ") {
				t.Fatalf("output lost Unicode label or target node %q:\n%s", label, output)
			}
		})
	}
}

func TestMaxWidthPreservesGraphContentAndDoesNotAffectSequence(t *testing.T) {
	graphConfig := diagram.NewTestConfig(true, "cli")
	graphConfig.MaxWidth = 20
	graphOutput, graphStatus, err := RenderDiagramWithStatus("graph LR\nA[LongLabel] --> B", graphConfig)
	if err != nil {
		t.Fatal(err)
	}
	for _, content := range []string{"LongLabel", "B"} {
		if !strings.Contains(graphOutput, content) {
			t.Fatalf("compact graph output lost %q:\n%s", content, graphOutput)
		}
	}
	if !graphStatus.Compacted || !graphStatus.Met {
		t.Fatalf("graph status = %+v, want compact success", graphStatus)
	}

	sequenceConfig := diagram.NewTestConfig(true, "cli")
	sequenceConfig.MaxWidth = 1
	sequenceOutput, sequenceStatus, err := RenderDiagramWithStatus("sequenceDiagram\nAlice->>Bob: Hello", sequenceConfig)
	if err != nil {
		t.Fatal(err)
	}
	if sequenceStatus.Requested || !strings.Contains(sequenceOutput, "Hello") {
		t.Fatalf("sequence width fitting should be ignored: status=%+v output=%s", sequenceStatus, sequenceOutput)
	}
}

func maxDisplayWidth(output string) int {
	width := 0
	for _, line := range strings.Split(output, "\n") {
		width = max(width, runewidth.StringWidth(line))
	}
	return width
}
