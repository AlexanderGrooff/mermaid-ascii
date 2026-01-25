package cmd

import "testing"

func TestGraphNodeLabelAlias(t *testing.T) {
	mermaid := "graph LR\nA[Alpha]\nA --> B\n"
	properties, err := mermaidFileToMap(mermaid, "cli")
	if err != nil {
		t.Fatalf("Failed to parse mermaid: %v", err)
	}

	if _, ok := properties.data.Get("A"); ok {
		t.Fatalf("expected node id 'A' to resolve to its label")
	}
	if _, ok := properties.data.Get("A[Alpha]"); ok {
		t.Fatalf("expected raw label syntax not to be treated as a node name")
	}
	if _, ok := properties.data.Get("Alpha"); !ok {
		t.Fatalf("expected label node 'Alpha' to exist")
	}
}

func TestSubgraphLabelAlias(t *testing.T) {
	mermaid := "graph LR\nsubgraph Foo[\"Group Label\"]\nA\nend\n"
	properties, err := mermaidFileToMap(mermaid, "cli")
	if err != nil {
		t.Fatalf("Failed to parse mermaid: %v", err)
	}
	if len(properties.subgraphs) != 1 {
		t.Fatalf("expected 1 subgraph, got %d", len(properties.subgraphs))
	}
	if properties.subgraphs[0].name != "Group Label" {
		t.Fatalf("expected subgraph label %q, got %q", "Group Label", properties.subgraphs[0].name)
	}
}
