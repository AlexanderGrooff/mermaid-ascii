package er

import (
	"strings"
	"testing"
)

// TestLayoutPlacement checks all entity boxes appear and don't overlap.
func TestLayoutPlacement(t *testing.T) {
	d, err := Parse("erDiagram\n CUSTOMER ||--o{ ORDER : places\n CUSTOMER {\n  int id PK\n }\n ORDER {\n  int id PK\n }\n LINE_ITEM {\n  int qty\n }")
	if err != nil {
		t.Fatal(err)
	}
	out := Render(d, false)
	for _, name := range []string{"CUSTOMER", "ORDER", "LINE_ITEM"} {
		if !strings.Contains(out, name) {
			t.Errorf("output missing entity %q:\n%s", name, out)
		}
	}
	// Boxes must not overlap: verify placement rectangles are disjoint.
	lay := placeEntities(d, unicodeGlyphs)
	placed := lay.placed
	for i := 0; i < len(placed); i++ {
		for j := i + 1; j < len(placed); j++ {
			a, b := placed[i], placed[j]
			if a.x < b.x+b.w && b.x < a.x+a.w && a.y < b.y+b.h && b.y < a.y+a.h {
				t.Errorf("boxes %q and %q overlap", a.entity.Name, b.entity.Name)
			}
		}
	}
}
