package er

import (
	"strings"
	"testing"
)

// TestRenderRelationship checks a relationship draws crow's-foot markers at both
// ends and the label, without corrupting the entity boxes.
func TestRenderRelationship(t *testing.T) {
	d, err := Parse("erDiagram\n CUSTOMER ||--o{ ORDER : places\n CUSTOMER {\n  int id PK\n }\n ORDER {\n  int id PK\n }")
	if err != nil {
		t.Fatal(err)
	}
	out := Render(d, false)
	for _, want := range []string{"CUSTOMER", "ORDER", "||", "o{", "places", "│ int │ id"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}
