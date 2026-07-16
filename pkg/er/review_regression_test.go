package er

import (
	"strings"
	"testing"
)

// TestUnquotedAliasForms covers mermaid's documented alias syntaxes: quoted
// (`a["Customer Account"]`), unquoted (`p[Person]`), and both as lone
// declarations and block headers.
func TestUnquotedAliasForms(t *testing.T) {
	d, err := Parse("erDiagram\n p[Person] {\n  string firstName\n }\n a[\"Customer Account\"] {\n  string email\n }\n p ||--o| a : has")
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Entities) != 2 {
		t.Fatalf("want 2 entities, got %d", len(d.Entities))
	}
	if d.Entities[0].Display != "Person" || d.Entities[1].Display != "Customer Account" {
		t.Errorf("aliases not applied: %q, %q", d.Entities[0].Display, d.Entities[1].Display)
	}
	if d.Relationships[0].Left != "p" || d.Relationships[0].Right != "a" {
		t.Errorf("relationship should reference ids p/a, got %q/%q",
			d.Relationships[0].Left, d.Relationships[0].Right)
	}

	// The lone (no-block) forms declare the same aliases.
	d2, err := Parse("erDiagram\n p[Person]\n a[\"Customer Account\"]")
	if err != nil {
		t.Fatal(err)
	}
	if d2.Entities[0].Display != "Person" || d2.Entities[1].Display != "Customer Account" {
		t.Errorf("lone aliases not applied: %q, %q", d2.Entities[0].Display, d2.Entities[1].Display)
	}
}

// TestQuotedEntityNameInRelationship checks a quoted multi-word entity name
// works on either side of a relationship line, matching its declaration.
func TestQuotedEntityNameInRelationship(t *testing.T) {
	d, err := Parse("erDiagram\n \"ORDER ITEM\" {\n  int id\n }\n CUSTOMER ||--o{ \"ORDER ITEM\" : has\n \"ORDER ITEM\" ||--|| SKU : tracks")
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Entities) != 3 {
		t.Fatalf("want 3 entities, got %d: quoted name split apart", len(d.Entities))
	}
	if d.Relationships[0].Right != "ORDER ITEM" || d.Relationships[1].Left != "ORDER ITEM" {
		t.Errorf("quoted name not matched to declaration: %+v", d.Relationships)
	}
}

// TestWideRuneAlignment checks double-width (CJK) names and labels keep box
// borders aligned: every border column of a box must line up across its rows.
func TestWideRuneAlignment(t *testing.T) {
	d, err := Parse("erDiagram\n 顧客 ||--o{ ORDER : 注文\n ORDER ||--|| LINE : has")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(Render(d, false), "\n")
	// The top border ┌…┐ of each box must sit exactly above its └…┘/├…┤ columns.
	top := lines[0]
	for _, col := range []rune{'┌', '┐'} {
		if !strings.ContainsRune(top, col) {
			t.Fatalf("no %c in top border: %q", col, top)
		}
	}
	openCols := runeColumns(top, '┌')
	closeCols := runeColumns(lines[2], '└', '┴', '├')
	for _, c := range openCols {
		if !containsInt(closeCols, c) {
			t.Errorf("box corner at column %d misaligned (CJK width bug):\n%s",
				c, strings.Join(lines[:4], "\n"))
		}
	}
}

// TestSelfLoopTokensIntact checks both cardinality tokens of a self-loop
// survive even on boxes with short names.
func TestSelfLoopTokensIntact(t *testing.T) {
	for _, name := range []string{"A", "ABC", "EMPLOYEE"} {
		d, err := Parse("erDiagram\n " + name + " ||--o{ " + name + " : manages")
		if err != nil {
			t.Fatal(err)
		}
		out := Render(d, false)
		if !strings.Contains(out, "||") || !strings.Contains(out, "o{") {
			t.Errorf("%s self-loop lost a token:\n%s", name, out)
		}
		if !strings.Contains(out, "manages") {
			t.Errorf("%s self-loop lost its label:\n%s", name, out)
		}
	}
}

// TestManyRelationshipsDistinctAttach checks each relationship gets its own
// attach column: n relationships on one pair produce n tees on each border.
func TestManyRelationshipsDistinctAttach(t *testing.T) {
	d, err := Parse("erDiagram\n A ||--o{ B : first\n A ||--o{ B : second\n B ||--|| A : third")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(Render(d, false), "\n")
	// Row 2 holds both bottom borders; each box needs 3 distinct ┬ tees.
	if n := strings.Count(lines[2], "┬"); n != 6 {
		t.Errorf("want 6 attach tees (3 per box), got %d:\n%s", n, lines[2])
	}
}

// TestEntityNamedClass checks style-directive keywords still work as entity
// names, in both block and relationship position.
func TestEntityNamedClass(t *testing.T) {
	d, err := Parse("erDiagram\n class {\n  int id\n }\n class ||--|| B : has")
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Entities) != 2 || len(d.Relationships) != 1 {
		t.Fatalf("entity named class mishandled: %d entities, %d relationships",
			len(d.Entities), len(d.Relationships))
	}
}

// TestParseErrorLineNumbers checks reported line numbers count comment lines.
func TestParseErrorLineNumbers(t *testing.T) {
	_, err := Parse("erDiagram\n %% one\n %% two\n A ||--|| B : ok\n x : y : z")
	if err == nil || !strings.Contains(err.Error(), "line 5") {
		t.Errorf("want error at line 5, got %v", err)
	}
}

// TestAccDescrBlockSkipped checks the multi-line accDescr { … } form is
// skipped without swallowing entities or erroring.
func TestAccDescrBlockSkipped(t *testing.T) {
	d, err := Parse("erDiagram\n accTitle: My title\n accDescr {\n a long\n description\n }\n A ||--|| B : has")
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Entities) != 2 {
		t.Fatalf("accDescr block leaked entities: got %d", len(d.Entities))
	}
}

// TestLongLabelSurvivesManyLanes checks a label longer than the lane count
// still renders whole (the gutter budgets lanes + label width).
func TestLongLabelSurvivesManyLanes(t *testing.T) {
	var b strings.Builder
	b.WriteString("erDiagram\n")
	for i := 0; i < 14; i++ {
		b.WriteString(strings.Repeat(" ", 1))
		b.WriteString(entityName(i%9) + " ||--|| " + entityName((i+1)%9) + " : r\n")
	}
	b.WriteString(" E0 ||--|| E7 : twenty-char-label-xx\n")
	d, err := Parse(b.String())
	if err != nil {
		t.Fatal(err)
	}
	if out := Render(d, false); !strings.Contains(out, "twenty-char-label-xx") {
		t.Errorf("long label truncated:\n%s", out)
	}
}

func entityName(i int) string { return "E" + string(rune('0'+i)) }

func runeColumns(s string, targets ...rune) []int {
	var cols []int
	col := 0
	for _, r := range s {
		for _, t := range targets {
			if r == t {
				cols = append(cols, col)
			}
		}
		col++
	}
	return cols
}

func containsInt(xs []int, x int) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}
