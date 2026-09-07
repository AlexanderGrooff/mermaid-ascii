package graph

import (
	"fmt"
	"strings"
	"testing"

	"github.com/AlexanderGrooff/mermaid-ascii/pkg/diagram"
	"github.com/mattn/go-runewidth"
)

func TestRenderGraphHandlesLongChainWithoutPanic(t *testing.T) {
	config := diagram.NewTestConfig(true, "cli")

	var b strings.Builder
	b.WriteString("graph TD\n")
	for i := 1; i < 30; i++ {
		fmt.Fprintf(&b, "N%d --> N%d\n", i, i+1)
	}

	output, err := renderGraph(b.String(), config)
	if err != nil {
		t.Fatalf("renderGraph() error = %v", err)
	}

	if !strings.Contains(output, "N30") {
		t.Fatalf("expected output to contain the last node\noutput:\n%s", output)
	}
}

func TestRenderGraphKeepsDisplayWidthForWideNodeLabels(t *testing.T) {
	config := diagram.NewTestConfig(true, "cli")
	output, err := renderGraph("graph LR\nA[\"中A\"] --> B", config)
	if err != nil {
		t.Fatalf("renderGraph() error = %v", err)
	}

	assertUniformDisplayWidth(t, output)
}

func TestRenderGraphKeepsDisplayWidthForWideSubgraphTitles(t *testing.T) {
	config := diagram.NewTestConfig(true, "cli")
	output, err := renderGraph("graph LR\nsubgraph sg [数据库]\nA --> B\nend", config)
	if err != nil {
		t.Fatalf("renderGraph() error = %v", err)
	}

	assertUniformDisplayWidth(t, output)
}

func TestRenderGraphKeepsExplicitTargetLabelAfterBareReference(t *testing.T) {
	config := diagram.NewTestConfig(true, "cli")
	output, err := renderGraph("graph TD\nA[\"Foo\"] --> B[\"Bar\"]\nB --> C[\"Baz\"]", config)
	if err != nil {
		t.Fatalf("renderGraph() error = %v", err)
	}

	if !strings.Contains(output, "Bar") {
		t.Fatalf("expected output to contain Bar\noutput:\n%s", output)
	}
	if strings.Contains(output, "\n|  B  |") || strings.Contains(output, "\n| B |\n") {
		t.Fatalf("expected B node to keep explicit label\noutput:\n%s", output)
	}
}

func TestRenderGraphKeepsStandaloneSubgraphLabelWhenReferencedLater(t *testing.T) {
	config := diagram.NewTestConfig(true, "cli")
	output, err := renderGraph("graph TD\nsubgraph one\n    A[\"VcpuManager\"]\nend\nA --> B", config)
	if err != nil {
		t.Fatalf("renderGraph() error = %v", err)
	}

	if !strings.Contains(output, "VcpuManager") {
		t.Fatalf("expected output to contain VcpuManager\noutput:\n%s", output)
	}
	if strings.Contains(output, "\n| A |\n") || strings.Contains(output, "\n|  A  |\n") {
		t.Fatalf("expected A node to keep standalone explicit label\noutput:\n%s", output)
	}
}

func TestRenderGraphSupportsLiteralNewlineInNodeLabel(t *testing.T) {
	config := diagram.NewTestConfig(true, "cli")
	output, err := renderGraph("graph LR\nA[\"line1\nline2\"] --> B", config)
	if err != nil {
		t.Fatalf("renderGraph() error = %v", err)
	}

	if !strings.Contains(output, "line1") || !strings.Contains(output, "line2") {
		t.Fatalf("expected output to contain both label lines\noutput:\n%s", output)
	}
	if strings.Contains(output, "A[\"line1") || strings.Contains(output, "line2\"]") {
		t.Fatalf("expected parser to keep literal newline inside the label\noutput:\n%s", output)
	}
}

func TestRenderGraphSeparatesDuplicateEdgeLabels(t *testing.T) {
	config := diagram.NewTestConfig(true, "cli")
	output, err := renderGraph("graph LR\nA -->|miss| B\nA -->|hit| B", config)
	if err != nil {
		t.Fatalf("renderGraph() error = %v", err)
	}

	if strings.Contains(output, "mhit") {
		t.Fatalf("expected duplicate edge labels not to merge\noutput:\n%s", output)
	}
	if !strings.Contains(output, "miss") || !strings.Contains(output, "hit") {
		t.Fatalf("expected output to contain both duplicate edge labels\noutput:\n%s", output)
	}

	missLine := -1
	hitLine := -1
	for i, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "miss") {
			missLine = i
		}
		if strings.Contains(line, "hit") {
			hitLine = i
		}
	}
	if missLine == -1 || hitLine == -1 || missLine == hitLine {
		t.Fatalf("expected duplicate edge labels on separate lines\noutput:\n%s", output)
	}
}

func TestRenderGraphSeparatesBidirectionalEdgeLabelsLR(t *testing.T) {
	config := diagram.NewTestConfig(true, "cli")
	output, err := renderGraph("graph LR\nA -->|workload exits| B\nB -->|run| A", config)
	if err != nil {
		t.Fatalf("renderGraph() error = %v", err)
	}

	if strings.Contains(output, "worklorunexits") {
		t.Fatalf("expected bidirectional edge labels not to merge\noutput:\n%s", output)
	}
	if !strings.Contains(output, "workload") || !strings.Contains(output, "exits") || !strings.Contains(output, "run") {
		t.Fatalf("expected output to contain both bidirectional edge labels\noutput:\n%s", output)
	}

	workloadLine := -1
	runLine := -1
	for i, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "workload") {
			workloadLine = i
		}
		if strings.Contains(line, "run") {
			runLine = i
		}
	}
	if workloadLine == -1 || runLine == -1 || workloadLine == runLine {
		t.Fatalf("expected bidirectional edge labels on separate lines\noutput:\n%s", output)
	}
}

func TestRenderGraphSeparatesBidirectionalEdgeLabelsTD(t *testing.T) {
	config := diagram.NewTestConfig(true, "cli")
	output, err := renderGraph("graph TD\nA -->|forward| B\nB -->|back| A", config)
	if err != nil {
		t.Fatalf("renderGraph() error = %v", err)
	}

	if strings.Contains(output, "fbackrd") {
		t.Fatalf("expected bidirectional edge labels not to merge\noutput:\n%s", output)
	}
	if !strings.Contains(output, "forward") || !strings.Contains(output, "back") {
		t.Fatalf("expected output to contain both bidirectional edge labels\noutput:\n%s", output)
	}
}

func TestRenderGraphAlignsFanoutLabels(t *testing.T) {
	cases := []struct {
		name   string
		labels []string
		tail   string
	}{
		{"three", []string{"alpha", "beta", "gamma"}, ""},
		{"five", []string{"alpha", "beta", "gamma", "delta", "epsilon"}, ""},
		{"wide", []string{"long first label", "猫猫猫", "last choice"}, ""},
		{"unlabeled branch", []string{"alpha", "", "a longer last label"}, ""},
		{"downstream", []string{"alpha", "beta", "gamma"}, "B --> X\nC --> Y\nD --> Z\n"},
	}
	for _, tc := range cases {
		for _, ascii := range []bool{true, false} {
			for _, padding := range []int{0, 1, 5, 12} {
				t.Run(fmt.Sprintf("%s/ascii=%v/padding=%d", tc.name, ascii, padding), func(t *testing.T) {
					config := diagram.NewTestConfig(ascii, "cli")
					config.PaddingBetweenX, config.PaddingBetweenY = padding, padding
					var source strings.Builder
					source.WriteString("flowchart TB\n")
					for i, label := range tc.labels {
						if label == "" {
							fmt.Fprintf(&source, "A --> %c\n", 'B'+i)
						} else {
							fmt.Fprintf(&source, "A -->|%s| %c\n", label, 'B'+i)
						}
					}
					source.WriteString(tc.tail)
					output, err := renderGraph(source.String(), config)
					if err != nil {
						t.Fatal(err)
					}
					assertUniformDisplayWidth(t, output)
					lines := strings.Split(output, "\n")
					labelRow := -1
					for i, label := range tc.labels {
						if label == "" {
							continue
						}
						if strings.Count(output, label) != 1 {
							t.Fatalf("expected label %q exactly once\n%s", label, output)
						}
						for row, line := range lines {
							index := strings.Index(line, label)
							if index < 0 {
								continue
							}
							if labelRow != -1 && labelRow != row {
								t.Fatalf("expected aligned labels\n%s", output)
							}
							labelRow = row
							center := runewidth.StringWidth(line[:index]) + runewidth.StringWidth(label)/2
							vertical, arrow, border := "|", "v", "|"
							if !ascii {
								vertical, arrow, border = "│", "▼", "│"
							}
							if row == 0 || row+2 >= len(lines) || string([]rune(lines[row-1])[center]) != vertical || string([]rune(lines[row+1])[center]) != vertical {
								t.Fatalf("label %q must sit on its branch with clearance\n%s", label, output)
							}
							nodeText := fmt.Sprintf("%s %c %s", border, 'B'+i, border)
							foundTarget := false
							for targetRow := row + 2; targetRow < len(lines); targetRow++ {
								if target := strings.Index(lines[targetRow], nodeText); target >= 0 {
									if runewidth.StringWidth(lines[targetRow][:target])+2 != center || string([]rune(lines[targetRow-3])[center]) != arrow {
										t.Fatalf("label %q must align with its target and arrowhead\n%s", label, output)
									}
									foundTarget = true
									break
								}
							}
							if !foundTarget {
								t.Fatalf("expected unchanged target box %q\n%s", nodeText, output)
							}
						}
					}
				})
			}
		}
	}
}

func assertUniformDisplayWidth(t *testing.T, output string) {
	t.Helper()

	lines := strings.Split(output, "\n")
	if len(lines) == 0 {
		t.Fatal("expected rendered output")
	}

	want := runewidth.StringWidth(lines[0])
	for i, line := range lines[1:] {
		if got := runewidth.StringWidth(line); got != want {
			t.Fatalf("line %d display width = %d, want %d\noutput:\n%s", i+2, got, want, output)
		}
	}
}

// renderGraph parses and draws src with config, the way
// render.GraphDiagram does.
func renderGraph(src string, config *diagram.Config) (string, error) {
	styleType := config.StyleType
	if styleType == "" {
		styleType = "cli"
	}
	p, err := Parse(src, styleType)
	if err != nil {
		return "", err
	}
	p.Apply(config)
	return Draw(p), nil
}
