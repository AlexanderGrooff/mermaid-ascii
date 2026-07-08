package cmd

import "testing"

func TestSplitGraphLines(t *testing.T) {
	input := "graph LR\\nA[\"line1\\nline2\"] --> B\\nC --> D"

	got := splitGraphLines(input)
	want := []string{"graph LR", `A["line1\nline2"] --> B`, "C --> D"}

	if len(got) != len(want) {
		t.Fatalf("line count = %d, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("line %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestSplitGraphLinesKeepsLiteralNewlineInsideNodeLabel(t *testing.T) {
	input := "graph LR\nA[\"line1\nline2\"] --> B\nC --> D"

	got := splitGraphLines(input)
	want := []string{"graph LR", "A[\"line1\nline2\"] --> B", "C --> D"}

	if len(got) != len(want) {
		t.Fatalf("line count = %d, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("line %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestParseNodeWithExplicitLabel(t *testing.T) {
	node := parseNode(`A["line1<br/>line2"]:::primary`)

	if node.name != "A" {
		t.Fatalf("name = %q, want %q", node.name, "A")
	}
	if node.styleClass != "primary" {
		t.Fatalf("styleClass = %q, want %q", node.styleClass, "primary")
	}
	if len(node.label.lines) != 2 {
		t.Fatalf("label lines = %d, want 2", len(node.label.lines))
	}
	if node.label.lines[0] != "line1" || node.label.lines[1] != "line2" {
		t.Fatalf("label lines = %#v, want [line1 line2]", node.label.lines)
	}
}

func TestMermaidFileToMapPreservesEscapedLabelNewlines(t *testing.T) {
	properties, err := mermaidFileToMap("graph LR\\nA[\"line1\\nline2\"] --> B", "cli")
	if err != nil {
		t.Fatalf("mermaidFileToMap() error = %v", err)
	}

	spec := properties.nodeSpecs["A"]
	if len(spec.label.lines) != 2 {
		t.Fatalf("label lines = %d, want 2", len(spec.label.lines))
	}
	if spec.label.lines[0] != "line1" || spec.label.lines[1] != "line2" {
		t.Fatalf("label lines = %#v, want [line1 line2]", spec.label.lines)
	}
}

func TestMermaidFileToMapPreservesLiteralLabelNewlines(t *testing.T) {
	properties, err := mermaidFileToMap("graph LR\nA[\"line1\nline2\"] --> B", "cli")
	if err != nil {
		t.Fatalf("mermaidFileToMap() error = %v", err)
	}

	spec := properties.nodeSpecs["A"]
	if len(spec.label.lines) != 2 {
		t.Fatalf("label lines = %d, want 2", len(spec.label.lines))
	}
	if spec.label.lines[0] != "line1" || spec.label.lines[1] != "line2" {
		t.Fatalf("label lines = %#v, want [line1 line2]", spec.label.lines)
	}
}

func TestParseSubgraphHeader(t *testing.T) {
	tests := []struct {
		name      string
		header    string
		wantID    string
		wantLabel string
	}{
		{
			name:      "plain title",
			header:    "Frontend",
			wantID:    "",
			wantLabel: "Frontend",
		},
		{
			name:      "explicit id and title",
			header:    "frontend [Frontend Services]",
			wantID:    "frontend",
			wantLabel: "Frontend Services",
		},
		{
			name:      "quoted title",
			header:    `frontend["Frontend Services"]`,
			wantID:    "frontend",
			wantLabel: "Frontend Services",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sg := parseSubgraphHeader(tt.header)
			if sg.id != tt.wantID {
				t.Fatalf("id = %q, want %q", sg.id, tt.wantID)
			}
			if sg.name != tt.wantLabel {
				t.Fatalf("name = %q, want %q", sg.name, tt.wantLabel)
			}
			if len(sg.label.lines) != 1 || sg.label.lines[0] != tt.wantLabel {
				t.Fatalf("label lines = %#v, want [%s]", sg.label.lines, tt.wantLabel)
			}
		})
	}
}

func TestMermaidFileToMapParsesSubgraphIDAndTitle(t *testing.T) {
	properties, err := mermaidFileToMap("graph LR\nsubgraph frontend [Frontend Services]\nA --> B\nend", "cli")
	if err != nil {
		t.Fatalf("mermaidFileToMap() error = %v", err)
	}

	if len(properties.subgraphs) != 1 {
		t.Fatalf("subgraphs = %d, want 1", len(properties.subgraphs))
	}

	sg := properties.subgraphs[0]
	if sg.id != "frontend" {
		t.Fatalf("id = %q, want %q", sg.id, "frontend")
	}
	if sg.name != "Frontend Services" {
		t.Fatalf("name = %q, want %q", sg.name, "Frontend Services")
	}
}

func TestMermaidFileToMapKeepsExplicitNodeLabelAcrossBareReferences(t *testing.T) {
	properties, err := mermaidFileToMap("graph TD\nA[\"Foo\"] --> B[\"Bar\"]\nB --> C[\"Baz\"]", "cli")
	if err != nil {
		t.Fatalf("mermaidFileToMap() error = %v", err)
	}

	spec := properties.nodeSpecs["B"]
	if len(spec.label.lines) != 1 || spec.label.lines[0] != "Bar" {
		t.Fatalf("label lines = %#v, want [Bar]", spec.label.lines)
	}
	if !spec.labelIsExplicit {
		t.Fatal("expected B label to remain explicit")
	}
}

func TestMermaidFileToMapUsesLatestExplicitLabel(t *testing.T) {
	properties, err := mermaidFileToMap("graph TD\nA[\"Old\"] --> B\nA[\"New\"] --> C", "cli")
	if err != nil {
		t.Fatalf("mermaidFileToMap() error = %v", err)
	}

	spec := properties.nodeSpecs["A"]
	if len(spec.label.lines) != 1 || spec.label.lines[0] != "New" {
		t.Fatalf("label lines = %#v, want [New]", spec.label.lines)
	}
	if !spec.labelIsExplicit {
		t.Fatal("expected A label to remain explicit")
	}
}

// TestGraphTypeDetection verifies that the diagram declaration line is parsed
// tolerantly: surrounding whitespace, a missing direction (defaults to
// top-down), and the reverse directions RL/BT are all accepted.
func TestGraphTypeDetection(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantDir string
		wantErr bool
	}{
		{"plain graph TD", "graph TD\nA --> B", "TD", false},
		{"flowchart LR", "flowchart LR\nA --> B", "LR", false},
		{"leading whitespace", "    flowchart LR\n    A --> B", "LR", false},
		{"trailing whitespace", "graph LR    \nA --> B", "LR", false},
		{"indented graph TD", "        graph TD\n        A --> B", "TD", false},
		{"bare graph defaults to TD", "graph\nA --> B", "TD", false},
		{"bare flowchart defaults to TD", "flowchart\nA --> B", "TD", false},
		{"TB maps to TD", "flowchart TB\nA --> B", "TD", false},
		{"RL maps to LR axis", "graph RL\nA --> B", "LR", false},
		{"BT maps to TD axis", "flowchart BT\nA --> B", "TD", false},
		{"trailing semicolon bare", "graph;\nA --> B", "TD", false},
		{"trailing semicolon with direction", "graph TD;\nA --> B", "TD", false},
		{"flowchart LR semicolon", "flowchart LR;\nA --> B", "LR", false},
		{"CRLF line ending", "graph TD\r\nA --> B", "TD", false},
		{"tab separator", "graph\tLR\nA --> B", "LR", false},
		{"lowercase direction errors", "graph td\nA --> B", "", true},
		{"extra tokens error", "graph TD foo\nA --> B", "", true},
		{"unknown type errors", "sequenceDiagram\nA->>B: x", "", true},
		{"unknown direction errors", "graph SIDEWAYS\nA --> B", "", true},
		{"empty input errors", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			props, err := mermaidFileToMap(tt.input, "cli")
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got direction %q", props.graphDirection)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if props.graphDirection != tt.wantDir {
				t.Errorf("direction = %q, want %q", props.graphDirection, tt.wantDir)
			}
		})
	}
}
