package er

import (
	"strings"
	"testing"
)

// TestRenderEntity checks the attribute-table rendering: name header, the
// shown columns, aligned content, and that empty columns are dropped.
func TestRenderEntity(t *testing.T) {
	d, err := Parse("erDiagram\n USER {\n  int id PK\n  string email UK \"unique\"\n  text bio\n }")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	out := strings.Join(renderEntity(d.Entities[0], unicodeGlyphs), "\n")

	for _, want := range []string{"USER", "id", "int", "PK", "email", "unique", "bio"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
	// A comment column is present (one attr has a comment); render is a bordered box.
	if !strings.HasPrefix(out, "┌") || !strings.HasSuffix(out, "┘") {
		t.Errorf("expected a bordered box:\n%s", out)
	}

	// No-comment, no-key entity drops those columns (only type+name).
	d2, _ := Parse("erDiagram\n T {\n  int a\n  int b\n }")
	out2 := strings.Join(renderEntity(d2.Entities[0], unicodeGlyphs), "\n")
	if strings.Count(out2, "┬") != 1 { // one column separator ⇒ two columns
		t.Errorf("expected exactly two columns (type,name):\n%s", out2)
	}
}
