package cmd

import (
	"strings"
	"testing"

	"github.com/AlexanderGrooff/mermaid-ascii/pkg/diagram"
)

// TestRenderWithFrontmatter checks frontmatter is stripped before type
// detection for every diagram type, and the title is printed above the output.
func TestRenderWithFrontmatter(t *testing.T) {
	cases := []struct {
		name      string
		input     string
		wantType  string
		wantParts []string
	}{
		{
			"sequence with title",
			"---\ntitle: Login flow\n---\nsequenceDiagram\nAlice->>Bob: Hi",
			"sequence",
			[]string{"Login flow", "Alice", "Bob", "Hi"},
		},
		{
			"er with theme config",
			"---\nconfig:\n  themeCSS: |\n    rect { fill: red; }\n---\nerDiagram\n  CUSTOMER ||--o{ ORDER : places",
			"er",
			[]string{"CUSTOMER", "ORDER", "places"},
		},
		{
			"graph with title",
			"---\ntitle: Deps\n---\ngraph LR\nA-->B",
			"graph",
			[]string{"Deps", "A", "B"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			stripped, _ := diagram.StripFrontmatter(c.input)
			diag, err := DiagramFactory(stripped)
			if err != nil {
				t.Fatal(err)
			}
			if diag.Type() != c.wantType {
				t.Errorf("detected %q, want %q", diag.Type(), c.wantType)
			}
			out, err := RenderDiagram(c.input, diagram.DefaultConfig())
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range c.wantParts {
				if !strings.Contains(out, want) {
					t.Errorf("output missing %q:\n%s", want, out)
				}
			}
		})
	}
}

// TestRenderTitleAboveDiagram checks the title sits on the first line,
// separated from the diagram by a blank line.
func TestRenderTitleAboveDiagram(t *testing.T) {
	out, err := RenderDiagram("---\ntitle: My title\n---\ngraph LR\nA-->B", diagram.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(out, "\n")
	if lines[0] != "My title" || lines[1] != "" {
		t.Errorf("want title + blank line first, got %q, %q", lines[0], lines[1])
	}
}
