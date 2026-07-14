package er

import "testing"

// TestCardinalityMatrix locks all 16 crow's-foot token combinations.
func TestCardinalityMatrix(t *testing.T) {
	lefts := map[string]Cardinality{"||": OnlyOne, "|o": ZeroOrOne, "}o": ZeroOrMore, "}|": OneOrMore}
	rights := map[string]Cardinality{"||": OnlyOne, "o|": ZeroOrOne, "o{": ZeroOrMore, "|{": OneOrMore}
	for lt, lc := range lefts {
		for rt, rc := range rights {
			in := "erDiagram\n A " + lt + "--" + rt + " B : r"
			d, err := Parse(in)
			if err != nil {
				t.Fatalf("%q: %v", in, err)
			}
			r := d.Relationships[0]
			if r.LeftCard != lc || r.RightCard != rc || !r.Identifying {
				t.Errorf("%q => L=%v R=%v ident=%v", in, r.LeftCard, r.RightCard, r.Identifying)
			}
		}
	}
}

// TestIdentifyingVariants covers --, .., -., .- and word aliases.
func TestIdentifyingVariants(t *testing.T) {
	cases := []struct {
		in    string
		ident bool
	}{
		{"erDiagram\n A ||--|| B : r", true},
		{"erDiagram\n A ||..|| B : r", false},
		{"erDiagram\n A ||-.|| B : r", false},
		{"erDiagram\n A ||.-|| B : r", false},
		{"erDiagram\n A only one to one or more B : r", true},         // word alias, identifying
		{"erDiagram\n A many optionally to zero or one B : r", false}, // word alias, non-identifying
	}
	for _, c := range cases {
		d, err := Parse(c.in)
		if err != nil {
			t.Fatalf("%q: %v", c.in, err)
		}
		if d.Relationships[0].Identifying != c.ident {
			t.Errorf("%q ident = %v, want %v", c.in, d.Relationships[0].Identifying, c.ident)
		}
	}
}

// TestTextAliasCardinality checks word/short cardinality aliases map correctly.
func TestTextAliasCardinality(t *testing.T) {
	d, err := Parse("erDiagram\n A many to one or more B : r")
	if err != nil {
		t.Fatal(err)
	}
	r := d.Relationships[0]
	if r.LeftCard != ZeroOrMore || r.RightCard != OneOrMore {
		t.Errorf("got L=%v R=%v, want ZeroOrMore, OneOrMore", r.LeftCard, r.RightCard)
	}
}

// TestParenTypesAndMultiBlock guards the paren-type + multi-block fixes.
func TestParenTypesAndMultiBlock(t *testing.T) {
	d, err := Parse("erDiagram\n T {\n  decimal(10, 2) amount\n }\n T {\n  varchar(255) name PK \"the name\"\n }")
	if err != nil {
		t.Fatal(err)
	}
	a := d.Entities[0].Attributes
	if len(a) != 2 {
		t.Fatalf("want 2 attrs (multi-block append), got %d", len(a))
	}
	if a[0].Type != "decimal(10, 2)" || a[0].Name != "amount" {
		t.Errorf("paren type parsed wrong: %+v", a[0])
	}
	if a[1].Type != "varchar(255)" || a[1].Name != "name" || len(a[1].Keys) != 1 || a[1].Comment != "the name" {
		t.Errorf("attr 2 wrong: %+v", a[1])
	}
}

// TestStyleLinesSkipped: styling/accessibility lines don't fail the diagram.
func TestStyleLinesSkipped(t *testing.T) {
	d, err := Parse("erDiagram\n A ||--|| B : r\n classDef foo fill:#f00\n class A foo\n style B color:#0f0\n accTitle: My ER\n")
	if err != nil {
		t.Fatalf("style lines should be skipped, got: %v", err)
	}
	if len(d.Relationships) != 1 {
		t.Errorf("want 1 relationship, got %d", len(d.Relationships))
	}
}

// TestRecursiveAndDuplicate: self-relationship + no duplicate entities.
func TestRecursiveAndDuplicate(t *testing.T) {
	d, err := Parse("erDiagram\n NODE ||--o{ NODE : parent\n NODE ||--|| NODE : self")
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Entities) != 1 {
		t.Errorf("recursive rels should not duplicate entity: got %d", len(d.Entities))
	}
	if len(d.Relationships) != 2 {
		t.Errorf("want 2 relationships, got %d", len(d.Relationships))
	}
}

// TestEitherSideCardinality: a token like o{ / }| may appear on the LEFT too.
func TestEitherSideCardinality(t *testing.T) {
	d, err := Parse("erDiagram\n grants o{--|| tx : has")
	if err != nil {
		t.Fatal(err)
	}
	r := d.Relationships[0]
	if r.LeftCard != ZeroOrMore || r.RightCard != OnlyOne {
		t.Errorf("got L=%v R=%v, want ZeroOrMore, OnlyOne", r.LeftCard, r.RightCard)
	}
}

// TestEntityAliases: block, bracket, and standalone alias forms.
func TestEntityAliases(t *testing.T) {
	cases := []struct{ in, id, display string }{
		{"erDiagram\n Signature draft {\n  int id\n }", "Signature", "draft"},
		{"erDiagram\n" + ` fua["Fresha User Account"] {` + "\n  int id\n }", "fua", "Fresha User Account"},
		{"erDiagram\n" + ` acct["Account Ledger"]`, "acct", "Account Ledger"},
	}
	for _, c := range cases {
		d, err := Parse(c.in)
		if err != nil {
			t.Fatalf("%q: %v", c.in, err)
		}
		e := d.Entities[0]
		if e.Name != c.id || e.Display != c.display {
			t.Errorf("%q => id=%q display=%q, want id=%q display=%q", c.in, e.Name, e.Display, c.id, c.display)
		}
	}
}

// TestDirectionAndLenientComments: direction directive, // notes, and an
// unclosed-quote comment are all tolerated.
func TestDirectionAndLenientComments(t *testing.T) {
	in := "erDiagram\n direction LR\n T {\n  // a note line\n  string s \"unclosed comment\n  int n\n }"
	d, err := Parse(in)
	if err != nil {
		t.Fatalf("expected tolerant parse, got: %v", err)
	}
	if len(d.Entities) != 1 || len(d.Entities[0].Attributes) != 2 {
		t.Errorf("want 1 entity with 2 attrs, got %d entities", len(d.Entities))
	}
}
