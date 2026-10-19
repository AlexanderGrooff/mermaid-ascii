package graph

import (
	"fmt"
	"testing"
)

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

func TestParseNodePreservesLiteralBoundaryQuote(t *testing.T) {
	node := parseNode(`A{literal"}`)

	if node.name != "A" {
		t.Fatalf("name = %q, want %q", node.name, "A")
	}
	if len(node.label.lines) != 1 || node.label.lines[0] != `literal"` {
		t.Fatalf("label lines = %#v, want %q", node.label.lines, `literal"`)
	}
	if !node.hasLabel {
		t.Fatal("expected shape declaration to have an explicit label")
	}
}

func TestParseNodeFallsBackForExtraClosingDelimiter(t *testing.T) {
	declaration := `A{bad}}`
	node := parseNode(declaration)

	if node.name != declaration {
		t.Fatalf("name = %q, want bare declaration %q", node.name, declaration)
	}
	if len(node.label.lines) != 1 || node.label.lines[0] != declaration {
		t.Fatalf("label lines = %#v, want [%s]", node.label.lines, declaration)
	}
	if node.hasLabel {
		t.Fatal("malformed shape should not be treated as an explicit label")
	}
}

func TestParseNodeFallsBackForUnmatchedQuoteBeforeClosingDelimiter(t *testing.T) {
	declaration := `A{bad\"}}`
	node := parseNode(declaration)

	if node.name != declaration {
		t.Fatalf("name = %q, want bare declaration %q", node.name, declaration)
	}
	if len(node.label.lines) != 1 || node.label.lines[0] != declaration {
		t.Fatalf("label lines = %#v, want [%s]", node.label.lines, declaration)
	}
	if node.hasLabel {
		t.Fatal("malformed shape should not be treated as an explicit label")
	}
}

func TestParseNodeShapes(t *testing.T) {
	tests := []struct {
		name        string
		declaration string
		wantName    string
		wantLabel   string
	}{
		{name: "braced", declaration: "A{Decision}", wantName: "A", wantLabel: "Decision"},
		{name: "parenthesized", declaration: "B(Round)", wantName: "B", wantLabel: "Round"},
		{name: "double braced", declaration: "C{{Hexagon}}", wantName: "C", wantLabel: "Hexagon"},
		{name: "cylindrical", declaration: "D[(Database)]", wantName: "D", wantLabel: "Database"},
		{name: "asymmetric", declaration: "E>Asymmetric]", wantName: "E", wantLabel: "Asymmetric"},
		{name: "square bracket", declaration: `F["Rectangle"]`, wantName: "F", wantLabel: "Rectangle"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := parseNode(tt.declaration)
			if node.name != tt.wantName {
				t.Fatalf("name = %q, want %q", node.name, tt.wantName)
			}
			if len(node.label.lines) != 1 || node.label.lines[0] != tt.wantLabel {
				t.Fatalf("label lines = %#v, want [%s]", node.label.lines, tt.wantLabel)
			}
			if !node.hasLabel {
				t.Fatal("expected shape declaration to have an explicit label")
			}
		})
	}
}

func TestMermaidFileToMapKeepsShapedNodeLabelsAcrossBareReferences(t *testing.T) {
	properties, err := Parse("graph LR\\nA{Decision}\\nB(Round)\\nC{{Hexagon}}\\nD[(Database)]\\nE>Asymmetric]\\nA --> B\\nB --> C\\nC --> D\\nD --> E", "cli")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	for _, node := range []struct {
		name  string
		label string
	}{
		{name: "A", label: "Decision"},
		{name: "B", label: "Round"},
		{name: "C", label: "Hexagon"},
		{name: "D", label: "Database"},
		{name: "E", label: "Asymmetric"},
	} {
		spec, ok := properties.nodeSpecs[node.name]
		if !ok {
			t.Fatalf("missing node spec for %s", node.name)
		}
		if len(spec.label.lines) != 1 || spec.label.lines[0] != node.label {
			t.Errorf("label for %s = %#v, want [%s]", node.name, spec.label.lines, node.label)
		}
		if !spec.labelIsExplicit {
			t.Errorf("label for %s should remain explicit", node.name)
		}
	}
}

func TestMermaidFileToMapPreservesEscapedLabelNewlines(t *testing.T) {
	properties, err := Parse("graph LR\\nA[\"line1\\nline2\"] --> B", "cli")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
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
	properties, err := Parse("graph LR\nA[\"line1\nline2\"] --> B", "cli")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
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
	properties, err := Parse("graph LR\nsubgraph frontend [Frontend Services]\nA --> B\nend", "cli")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
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
	properties, err := Parse("graph TD\nA[\"Foo\"] --> B[\"Bar\"]\nB --> C[\"Baz\"]", "cli")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
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
	properties, err := Parse("graph TD\nA[\"Old\"] --> B\nA[\"New\"] --> C", "cli")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	spec := properties.nodeSpecs["A"]
	if len(spec.label.lines) != 1 || spec.label.lines[0] != "New" {
		t.Fatalf("label lines = %#v, want [New]", spec.label.lines)
	}
	if !spec.labelIsExplicit {
		t.Fatal("expected A label to remain explicit")
	}
}

// Tests non-standard arrow operators.
func TestMermaidFileToMapParsesNonStandardEdgeOperators(t *testing.T) {
	for _, op := range []string{"-->", "-.->", "==>", "---", "--o", "--x"} {
		t.Run(op, func(t *testing.T) {
			properties, err := Parse("graph LR\nA "+op+" B", "cli")
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			edges, ok := properties.data.Get("A")
			if !ok || len(edges) != 1 || edges[0].child.name != "B" {
				t.Fatalf("edges from A = %#v, want one edge to B", edges)
			}
		})
	}
}

func TestMermaidFileToMapParsesChainedNonStandardEdges(t *testing.T) {
	input := "graph LR\nA -.-> B\nB ==> C\nC --- D\nD --o E\nE --x F"
	properties, err := Parse(input, "cli")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	for _, link := range [][2]string{{"A", "B"}, {"B", "C"}, {"C", "D"}, {"D", "E"}, {"E", "F"}} {
		edges, ok := properties.data.Get(link[0])
		if !ok || len(edges) != 1 || edges[0].child.name != link[1] {
			t.Fatalf("edges from %q = %#v, want one edge to %q", link[0], edges, link[1])
		}
	}
}

func TestParsePreservesQuotedLabelMarkersAndUnquotedClassShorthand(t *testing.T) {
	properties, err := Parse(`graph LR
A["literal %% and ::: text"]:::primary --> B`, "cli")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	spec := properties.nodeSpecs["A"]
	if len(spec.label.lines) != 1 || spec.label.lines[0] != "literal %% and ::: text" {
		t.Fatalf("label lines = %#v, want [literal %% and ::: text]", spec.label.lines)
	}
	if spec.styleClass != "primary" {
		t.Fatalf("styleClass = %q, want %q", spec.styleClass, "primary")
	}
}

func TestParseNodeDoesNotTreatQuotedClassMarkerAsShorthand(t *testing.T) {
	node := parseNode(`A["literal ::: text"]`)

	if node.name != "A" {
		t.Fatalf("name = %q, want %q", node.name, "A")
	}
	if node.styleClass != "" {
		t.Fatalf("styleClass = %q, want empty class", node.styleClass)
	}
	if len(node.label.lines) != 1 || node.label.lines[0] != "literal ::: text" {
		t.Fatalf("label lines = %#v, want [literal ::: text]", node.label.lines)
	}
}

func TestParseRejectsMalformedClassDefWithoutPanicking(t *testing.T) {
	for _, declaration := range []string{
		"classDef broken fill",
		"classDef broken fill:#f00,stroke",
		"classDef broken fill:#f00,",
	} {
		t.Run(declaration, func(t *testing.T) {
			if _, err := Parse("graph LR\\n"+declaration, "cli"); err == nil {
				t.Fatal("Parse() succeeded for malformed classDef")
			}
		})
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
		{"uppercase graph keyword errors", "GRAPH TD\nA --> B", "", true}, // flowchart stays case-sensitive (mermaid parity)
		{"extra tokens error", "graph TD foo\nA --> B", "", true},
		{"unknown type errors", "sequenceDiagram\nA->>B: x", "", true},
		{"unknown direction errors", "graph SIDEWAYS\nA --> B", "", true},
		{"empty input errors", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			props, err := Parse(tt.input, "cli")
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

func TestParsePreservesAmpersandInsideNodeLabel(t *testing.T) {
	input := `flowchart TD
A["Stage 1: Track & Commit Canonical Profiles"] --> B["Stage 2: Author Skill"]`
	properties, err := Parse(input, "cli")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(properties.nodeSpecs) != 2 {
		t.Fatalf("expected 2 node specs, got %d: %#v", len(properties.nodeSpecs), properties.nodeSpecs)
	}
	specA, ok := properties.nodeSpecs["A"]
	if !ok {
		t.Fatalf("expected node A in nodeSpecs")
	}
	if len(specA.label.lines) != 1 || specA.label.lines[0] != "Stage 1: Track & Commit Canonical Profiles" {
		t.Fatalf("node A label = %#v, want [Stage 1: Track & Commit Canonical Profiles]", specA.label.lines)
	}
	specB, ok := properties.nodeSpecs["B"]
	if !ok {
		t.Fatalf("expected node B in nodeSpecs")
	}
	if len(specB.label.lines) != 1 || specB.label.lines[0] != "Stage 2: Author Skill" {
		t.Fatalf("node B label = %#v, want [Stage 2: Author Skill]", specB.label.lines)
	}
	edges, ok := properties.data.Get("A")
	if !ok || len(edges) != 1 || edges[0].child.name != "B" {
		t.Fatalf("edges from A = %#v, want one edge to B", edges)
	}
}

func TestParseAmpersandEdgeCases(t *testing.T) {
	t.Run("ampersand inside unquoted shapes with outer fanout", func(t *testing.T) {
		shapes := []struct {
			name   string
			declA  string
			declB  string
			labelA string
			labelB string
		}{
			{name: "brackets", declA: "A[Track & Commit]", declB: "B[Build & Deploy]", labelA: "Track & Commit", labelB: "Build & Deploy"},
			{name: "parens", declA: "A(Track & Commit)", declB: "B(Build & Deploy)", labelA: "Track & Commit", labelB: "Build & Deploy"},
			{name: "braces", declA: "A{Track & Commit}", declB: "B{Build & Deploy}", labelA: "Track & Commit", labelB: "Build & Deploy"},
			{name: "hexagon", declA: "A{{Track & Commit}}", declB: "B{{Build & Deploy}}", labelA: "Track & Commit", labelB: "Build & Deploy"},
			{name: "cylinder", declA: "A[(Track & Commit)]", declB: "B[(Build & Deploy)]", labelA: "Track & Commit", labelB: "Build & Deploy"},
			{name: "asymmetric", declA: "A>Track & Commit]", declB: "B>Build & Deploy]", labelA: "Track & Commit", labelB: "Build & Deploy"},
		}

		for _, tc := range shapes {
			t.Run(tc.name, func(t *testing.T) {
				input := fmt.Sprintf("graph LR\n%s & %s --> C", tc.declA, tc.declB)
				props, err := Parse(input, "cli")
				if err != nil {
					t.Fatalf("Parse() error = %v", err)
				}
				specA, ok := props.nodeSpecs["A"]
				if !ok || len(specA.label.lines) != 1 || specA.label.lines[0] != tc.labelA {
					t.Fatalf("A spec = %#v, want label %q", specA, tc.labelA)
				}
				specB, ok := props.nodeSpecs["B"]
				if !ok || len(specB.label.lines) != 1 || specB.label.lines[0] != tc.labelB {
					t.Fatalf("B spec = %#v, want label %q", specB, tc.labelB)
				}
				edgesA, okA := props.data.Get("A")
				edgesB, okB := props.data.Get("B")
				if !okA || len(edgesA) != 1 || edgesA[0].child.name != "C" {
					t.Fatalf("edges from A = %#v, want edge to C", edgesA)
				}
				if !okB || len(edgesB) != 1 || edgesB[0].child.name != "C" {
					t.Fatalf("edges from B = %#v, want edge to C", edgesB)
				}
			})
		}
	})

	t.Run("ampersand on RHS with fan-out", func(t *testing.T) {
		input := "graph LR\nA --> B & C[\"Foo & Bar\"]"
		props, err := Parse(input, "cli")
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		specC, ok := props.nodeSpecs["C"]
		if !ok || len(specC.label.lines) != 1 || specC.label.lines[0] != "Foo & Bar" {
			t.Fatalf("C spec = %#v, want label 'Foo & Bar'", specC)
		}
		edgesA, okA := props.data.Get("A")
		if !okA || len(edgesA) != 2 {
			t.Fatalf("edges from A = %#v, want 2 edges", edgesA)
		}
	})

	t.Run("bare node name containing ampersand without spaces", func(t *testing.T) {
		input := "graph LR\nAT&T & R&D --> Verizon"
		props, err := Parse(input, "cli")
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		edgesATT, ok1 := props.data.Get("AT&T")
		edgesRD, ok2 := props.data.Get("R&D")
		if !ok1 || len(edgesATT) != 1 || edgesATT[0].child.name != "Verizon" {
			t.Fatalf("edges from AT&T = %#v, want edge to Verizon", edgesATT)
		}
		if !ok2 || len(edgesRD) != 1 || edgesRD[0].child.name != "Verizon" {
			t.Fatalf("edges from R&D = %#v, want edge to Verizon", edgesRD)
		}
	})

	t.Run("escaped quotes in node label containing ampersand", func(t *testing.T) {
		input := `graph LR
A["He said \"Fish & Chips\""] --> B`
		props, err := Parse(input, "cli")
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		specA, ok := props.nodeSpecs["A"]
		if !ok || len(specA.label.lines) != 1 || specA.label.lines[0] != `He said \"Fish & Chips\"` {
			t.Fatalf("A spec = %#v, want label with escaped quotes", specA)
		}
	})

	t.Run("bidirectional arrow with ampersand labels", func(t *testing.T) {
		input := `graph LR
A["Left & Right"] <--> B["North & South"]`
		props, err := Parse(input, "cli")
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		specA := props.nodeSpecs["A"]
		specB := props.nodeSpecs["B"]
		if specA.label.lines[0] != "Left & Right" || specB.label.lines[0] != "North & South" {
			t.Fatalf("labels mismatch: A=%v, B=%v", specA.label.lines, specB.label.lines)
		}
	})

	t.Run("edge label containing ampersand", func(t *testing.T) {
		input := `graph LR
A -->|Save & Exit| B`
		props, err := Parse(input, "cli")
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		edgesA, ok := props.data.Get("A")
		if !ok || len(edgesA) != 1 || edgesA[0].label != "Save & Exit" {
			t.Fatalf("edges from A = %#v, want label 'Save & Exit'", edgesA)
		}
	})

	t.Run("standalone node declaration with ampersand", func(t *testing.T) {
		input := `flowchart TD
A["Phase 1: Track & Commit"]`
		props, err := Parse(input, "cli")
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		if len(props.nodeSpecs) != 1 {
			t.Fatalf("expected 1 node, got %d", len(props.nodeSpecs))
		}
		specA, ok := props.nodeSpecs["A"]
		if !ok || specA.label.lines[0] != "Phase 1: Track & Commit" {
			t.Fatalf("A spec = %#v", specA)
		}
	})

	t.Run("multiple ampersand chain", func(t *testing.T) {
		input := "graph LR\nA & B & C & D --> E"
		props, err := Parse(input, "cli")
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		for _, name := range []string{"A", "B", "C", "D"} {
			edges, ok := props.data.Get(name)
			if !ok || len(edges) != 1 || edges[0].child.name != "E" {
				t.Fatalf("edges from %s = %#v, want 1 edge to E", name, edges)
			}
		}
	})

	t.Run("quoted shapes with fanout on LHS and RHS", func(t *testing.T) {
		input := `graph LR
A["Track & Commit"] & B["Build & Deploy"] --> C["Deploy & Run"] & D["Alert & Oncall"]`
		props, err := Parse(input, "cli")
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		for _, name := range []string{"A", "B"} {
			edges, ok := props.data.Get(name)
			if !ok || len(edges) != 2 {
				t.Fatalf("edges from %s = %#v, want 2 edges", name, edges)
			}
			destMap := map[string]bool{}
			for _, e := range edges {
				destMap[e.child.name] = true
			}
			if !destMap["C"] || !destMap["D"] {
				t.Fatalf("edges from %s do not contain C and D: %#v", name, edges)
			}
		}
		specA := props.nodeSpecs["A"]
		if specA.label.lines[0] != "Track & Commit" {
			t.Fatalf("specA label = %v", specA.label.lines)
		}
		specC := props.nodeSpecs["C"]
		if specC.label.lines[0] != "Deploy & Run" {
			t.Fatalf("specC label = %v", specC.label.lines)
		}
	})

	t.Run("RHS multi-ampersand chain", func(t *testing.T) {
		input := "graph LR\nA --> B & C & D"
		props, err := Parse(input, "cli")
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		edgesA, ok := props.data.Get("A")
		if !ok || len(edgesA) != 3 {
			t.Fatalf("edges from A = %#v, want 3 edges", edgesA)
		}
	})

	t.Run("mixed shapes with ampersand in fanout chain", func(t *testing.T) {
		input := `graph LR
A[Track & Commit] & B(Author & Skill) & C{Review & Approve} & D>Qux & Test] --> E`
		props, err := Parse(input, "cli")
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		for _, name := range []string{"A", "B", "C", "D"} {
			edges, ok := props.data.Get(name)
			if !ok || len(edges) != 1 || edges[0].child.name != "E" {
				t.Fatalf("edges from %s = %#v, want 1 edge to E", name, edges)
			}
		}
	})

	t.Run("nested angle brackets inside node label", func(t *testing.T) {
		input := "graph LR\nA[x > y & z] & B --> C"
		props, err := Parse(input, "cli")
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		specA := props.nodeSpecs["A"]
		if specA.label.lines[0] != "x > y & z" {
			t.Fatalf("specA label = %v, want 'x > y & z'", specA.label.lines)
		}
		edgesA, okA := props.data.Get("A")
		edgesB, okB := props.data.Get("B")
		if !okA || len(edgesA) != 1 || edgesA[0].child.name != "C" {
			t.Fatalf("edges from A = %#v, want edge to C", edgesA)
		}
		if !okB || len(edgesB) != 1 || edgesB[0].child.name != "C" {
			t.Fatalf("edges from B = %#v, want edge to C", edgesB)
		}
	})

	t.Run("subgraph fanout with ampersand in labels", func(t *testing.T) {
		input := `graph TD
subgraph S[Service Group]
A["Auth & Security"] & B["Billing & Payments"]
end
A --> C`
		props, err := Parse(input, "cli")
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		if len(props.subgraphs) != 1 {
			t.Fatalf("expected 1 subgraph, got %d", len(props.subgraphs))
		}
		sg := props.subgraphs[0]
		foundA, foundB := false, false
		for _, n := range sg.nodes {
			if n == "A" {
				foundA = true
			}
			if n == "B" {
				foundB = true
			}
		}
		if !foundA || !foundB {
			t.Fatalf("subgraph nodes = %#v, want A and B", sg.nodes)
		}
		specA := props.nodeSpecs["A"]
		if specA.label.lines[0] != "Auth & Security" {
			t.Fatalf("specA label = %v", specA.label.lines)
		}
		specB := props.nodeSpecs["B"]
		if specB.label.lines[0] != "Billing & Payments" {
			t.Fatalf("specB label = %v", specB.label.lines)
		}
	})
}
