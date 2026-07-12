package er

import "testing"

func TestIsErDiagram(t *testing.T) {
	for _, c := range []struct {
		in   string
		want bool
	}{
		{"erDiagram\n A ||--|| B : x", true},
		{"ERDIAGRAM\n A ||--|| B : x", true}, // mermaid er grammar is case-insensitive
		{"erDiagramFoo\n A", false},          // token boundary
		{"graph TD\n A-->B", false},
		{"sequenceDiagram\n A->>B: x", false},
		{"", false},
	} {
		if got := IsErDiagram(c.in); got != c.want {
			t.Errorf("IsErDiagram(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestParseRelationshipCardinalities(t *testing.T) {
	tests := []struct {
		token     string
		wantLeft  Cardinality
		wantRight Cardinality
		wantIdent bool
	}{
		{"||--||", OnlyOne, OnlyOne, true},
		{"||--o{", OnlyOne, ZeroOrMore, true},
		{"}|--|{", OneOrMore, OneOrMore, true},
		{"|o..o|", ZeroOrOne, ZeroOrOne, false}, // non-identifying (dashed)
		{"}o..|{", ZeroOrMore, OneOrMore, false},
	}
	for _, tt := range tests {
		t.Run(tt.token, func(t *testing.T) {
			d, err := Parse("erDiagram\n A " + tt.token + " B : rel")
			if err != nil {
				t.Fatalf("parse %q: %v", tt.token, err)
			}
			if len(d.Relationships) != 1 {
				t.Fatalf("want 1 relationship, got %d", len(d.Relationships))
			}
			r := d.Relationships[0]
			if r.LeftCard != tt.wantLeft || r.RightCard != tt.wantRight || r.Identifying != tt.wantIdent {
				t.Errorf("got L=%v R=%v ident=%v, want L=%v R=%v ident=%v",
					r.LeftCard, r.RightCard, r.Identifying, tt.wantLeft, tt.wantRight, tt.wantIdent)
			}
			if r.Left != "A" || r.Right != "B" || r.Label != "rel" {
				t.Errorf("endpoints/label wrong: %+v", r)
			}
		})
	}
}

func TestParseEntityAttributes(t *testing.T) {
	d, err := Parse("erDiagram\n USER {\n  int id PK\n  string email UK \"unique\"\n  bigint org_id FK\n  text bio\n }")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(d.Entities) != 1 {
		t.Fatalf("want 1 entity, got %d", len(d.Entities))
	}
	got := d.Entities[0].Attributes
	want := []Attribute{
		{Type: "int", Name: "id", Keys: []string{"PK"}},
		{Type: "string", Name: "email", Keys: []string{"UK"}, Comment: "unique"},
		{Type: "bigint", Name: "org_id", Keys: []string{"FK"}},
		{Type: "text", Name: "bio"},
	}
	if len(got) != len(want) {
		t.Fatalf("attrs = %d, want %d", len(got), len(want))
	}
	for i, w := range want {
		g := got[i]
		if g.Type != w.Type || g.Name != w.Name || g.Comment != w.Comment || len(g.Keys) != len(w.Keys) {
			t.Errorf("attr %d = %+v, want %+v", i, g, w)
		}
	}
}

func TestParseErErrors(t *testing.T) {
	for _, c := range []struct{ name, in, wantErr string }{
		{"missing keyword", "A ||--|| B : x", "expected"},
		{"unclosed block", "erDiagram\n A {\n  int id", "unclosed"},
		{"bad attribute", "erDiagram\n A {\n  oneword\n }", "type and name"},
		{"no entities", "erDiagram\n", "no entities"},
	} {
		t.Run(c.name, func(t *testing.T) {
			if _, err := Parse(c.in); err == nil {
				t.Errorf("expected error containing %q", c.wantErr)
			}
		})
	}
}
