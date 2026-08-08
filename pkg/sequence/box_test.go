package sequence

import (
	"strings"
	"testing"

	"github.com/AlexanderGrooff/mermaid-ascii/pkg/diagram"
)

// TestBoxParsing mirrors mermaid's box spec cases ('should handle box',
// 'box without color', 'box without description') plus the color forms found
// in real diagrams. The fill color has no ASCII meaning: it is parsed away so
// it never leaks into the title, then discarded.
func TestBoxParsing(t *testing.T) {
	cases := []struct {
		name      string
		boxLine   string
		wantTitle string
	}{
		{"color and title", "box green Group 1", "Group 1"},
		{"title only", "box Group 1", "Group 1"},
		{"color only", "box aqua", ""},
		{"transparent", "box transparent Group 1", "Group 1"},
		{"rgb with spaces and br title", "box rgb(23, 124, 207) <br>PROFESSIONAL PROFILES<br>", "PROFESSIONAL PROFILES"},
		{"named color then title", "box Green New serwis", "New serwis"},
		{"hex color", "box #ff0000 Alerts", "Alerts"},
		{"title that is not a color", "box Facade", "Facade"},
		{"bare box", "box", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d, err := Parse("sequenceDiagram\n" + c.boxLine + "\nparticipant a as Alice\nparticipant b as Bob\nend\nparticipant c as Charlie\na->>b: hi")
			if err != nil {
				t.Fatal(err)
			}
			if len(d.Boxes) != 1 {
				t.Fatalf("want 1 box, got %d", len(d.Boxes))
			}
			b := d.Boxes[0]
			if b.Title != c.wantTitle {
				t.Errorf("title = %q, want %q", b.Title, c.wantTitle)
			}
			if b.First != 0 || b.Last != 1 {
				t.Errorf("box spans participants %d..%d, want 0..1", b.First, b.Last)
			}
			if len(d.Participants) != 3 {
				t.Errorf("want 3 participants (Charlie outside the box), got %d", len(d.Participants))
			}
		})
	}
}

// TestBoxWithActors: actor declarations are valid inside a box.
func TestBoxWithActors(t *testing.T) {
	d, err := Parse("sequenceDiagram\nbox Team\nactor a as Alice\nparticipant b as Bob\nend\na->>b: hi")
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Boxes) != 1 || d.Boxes[0].First != 0 || d.Boxes[0].Last != 1 {
		t.Fatalf("box not recorded over actor+participant: %+v", d.Boxes)
	}
}

// TestMultipleBoxes: two sibling groups, each framing its own participants.
func TestMultipleBoxes(t *testing.T) {
	d, err := Parse("sequenceDiagram\nbox One\nparticipant a\nend\nbox Two\nparticipant b\nparticipant c\nend\na->>b: hi")
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Boxes) != 2 {
		t.Fatalf("want 2 boxes, got %d", len(d.Boxes))
	}
	if d.Boxes[0].First != 0 || d.Boxes[0].Last != 0 || d.Boxes[1].First != 1 || d.Boxes[1].Last != 2 {
		t.Errorf("box ranges wrong: %+v %+v", d.Boxes[0], d.Boxes[1])
	}
}

// TestBoxErrors: mermaid's grammar allows only participant declarations
// inside a box, forbids nesting, and requires end.
func TestBoxErrors(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"message inside box", "sequenceDiagram\nbox X\nparticipant a\na->>a: hi\nend", "only participant declarations"},
		{"fragment inside box", "sequenceDiagram\nbox X\nparticipant a\nloop retry\nend\nend", "only participant declarations"},
		{"nested box", "sequenceDiagram\nbox X\nbox Y\nparticipant a\nend\nend", "cannot nest"},
		{"unclosed box", "sequenceDiagram\nbox X\nparticipant a\na->>a: hi", "only participant declarations"},
		{"unclosed box at EOF", "sequenceDiagram\nbox X\nparticipant a", "unclosed box"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := Parse(c.in)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("want error containing %q, got %v", c.want, err)
			}
		})
	}
}

// TestEmptyBoxDropped: a box with no participants frames nothing and is
// dropped rather than rendered as a floating frame.
func TestEmptyBoxDropped(t *testing.T) {
	d, err := Parse("sequenceDiagram\nbox Empty\nend\nparticipant a\na->>a: hi")
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Boxes) != 0 {
		t.Errorf("empty box should be dropped, got %+v", d.Boxes)
	}
}

// TestBoxRenderStructure: every body line sits inside the box borders, the
// title is embedded whole in the top border, and a message crossing the
// border renders a crossing, not a hole.
func TestBoxRenderStructure(t *testing.T) {
	d, err := Parse("sequenceDiagram\nbox Payments Group\nparticipant GW\nparticipant PAY\nend\nparticipant S as Shop\nGW->>PAY: auth\nGW->>S: receipt")
	if err != nil {
		t.Fatal(err)
	}
	out, err := Render(d, diagram.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if !strings.Contains(lines[0], "─ Payments Group ") {
		t.Errorf("title missing from top border: %q", lines[0])
	}
	if !strings.HasPrefix(lines[0], "┌") || !strings.HasPrefix(lines[len(lines)-1], "└") {
		t.Errorf("missing box corners: first %q last %q", lines[0], lines[len(lines)-1])
	}
	// Every line between top and bottom border starts with the box vertical.
	for i := 1; i < len(lines)-1; i++ {
		if !strings.HasPrefix(lines[i], "│") {
			t.Errorf("line %d not inside box: %q", i, lines[i])
		}
	}
	// The receipt arrow leaves the box: its row must cross the border with ┼.
	found := false
	for _, l := range lines {
		if strings.Contains(l, "receipt") || strings.Contains(l, "►") {
			if strings.Contains(l, "┼") {
				found = true
			}
		}
	}
	if !found {
		t.Error("arrow crossing the box border should render ┼")
	}
}
