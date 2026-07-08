package sequence

import "testing"

// TestArrowTypes checks that each message arrow syntax parses to the correct
// ArrowType, including line style (solid/dotted) and whether it has a head.
func TestArrowTypes(t *testing.T) {
	tests := []struct {
		arrow      string
		want       ArrowType
		wantDotted bool
		wantHead   bool
	}{
		{"->>", SolidArrow, false, true},
		{"-->>", DottedArrow, true, true},
		{"->", SolidOpen, false, false},
		{"-->", DottedOpen, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.arrow, func(t *testing.T) {
			sd, err := Parse("sequenceDiagram\n A" + tt.arrow + "B: msg")
			if err != nil {
				t.Fatalf("parse %q: %v", tt.arrow, err)
			}
			if len(sd.Messages) != 1 {
				t.Fatalf("expected 1 message, got %d", len(sd.Messages))
			}
			got := sd.Messages[0].ArrowType
			if got != tt.want {
				t.Errorf("ArrowType = %v, want %v", got, tt.want)
			}
			if got.isDotted() != tt.wantDotted {
				t.Errorf("isDotted() = %v, want %v", got.isDotted(), tt.wantDotted)
			}
			if got.hasHead() != tt.wantHead {
				t.Errorf("hasHead() = %v, want %v", got.hasHead(), tt.wantHead)
			}
		})
	}
}

// TestUnsupportedArrowsRejected documents that arrow types we don't yet support
// (async -x/-), cross, and bidirectional <<->>) are rejected rather than
// silently mis-parsed. mermaid supports these; adding them is future work.
func TestUnsupportedArrowsRejected(t *testing.T) {
	for _, in := range []string{
		"sequenceDiagram\n A-xB: cross",
		"sequenceDiagram\n A-)B: async",
		"sequenceDiagram\n A--xB: dotted cross",
		"sequenceDiagram\n A--)B: dotted async",
		"sequenceDiagram\n A<<->>B: bidirectional",
	} {
		if _, err := Parse(in); err == nil {
			t.Errorf("expected error for unsupported arrow in %q, got none", in)
		}
	}
}

// TestOpenArrowEmptyLabel checks an open arrow with an empty label parses, like
// the existing ->> empty-label case.
func TestOpenArrowEmptyLabel(t *testing.T) {
	sd, err := Parse("sequenceDiagram\n A->B: ")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(sd.Messages) != 1 || sd.Messages[0].ArrowType != SolidOpen {
		t.Fatalf("expected 1 SolidOpen message, got %d msgs", len(sd.Messages))
	}
	if sd.Messages[0].Label != "" {
		t.Errorf("expected empty label, got %q", sd.Messages[0].Label)
	}
}
