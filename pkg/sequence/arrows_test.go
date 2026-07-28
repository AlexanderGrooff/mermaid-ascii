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
			if _, hasHead := got.head(Unicode, true); hasHead != tt.wantHead {
				t.Errorf("head() present = %v, want %v", hasHead, tt.wantHead)
			}
		})
	}
}

// TestArrowSpecParityForms pins the exact input forms mermaid's own test
// suites use for these arrows (sequenceDiagram.spec.js and the cypress
// rendering spec): no spaces around the arrow or after the colon, punctuation
// in labels, asymmetric spacing, and both message directions.
func TestArrowSpecParityForms(t *testing.T) {
	tests := []struct {
		in   string
		want ArrowType
	}{
		{"sequenceDiagram\nAlice-xBob:Hello Bob, how are you?", SolidCross},
		{"sequenceDiagram\nAlice--xBob:Hello Bob, how are you?", DottedCross},
		{"sequenceDiagram\nAlice-)Bob:Hello Bob, how are you?", SolidPoint},
		{"sequenceDiagram\nAlice--)Bob:Hello Bob, how are you?", DottedPoint},
		{"sequenceDiagram\nAlice<<->>Bob:Hello Bob, how are you?", BidirectionalSolid},
		{"sequenceDiagram\nAlice<<-->>Bob:Hello Bob, how are you?", BidirectionalDotted},
		{"sequenceDiagram\n Bob--x Alice: I am good thanks!", DottedCross},
		{"sequenceDiagram\n Alice --x Ola1: Bye!", DottedCross},
		{"sequenceDiagram\n John<<->>Alice: This also works the other way", BidirectionalSolid},
	}
	for _, tt := range tests {
		sd, err := Parse(tt.in)
		if err != nil {
			t.Errorf("Parse(%q): %v", tt.in, err)
			continue
		}
		if len(sd.Messages) != 1 || sd.Messages[0].ArrowType != tt.want {
			t.Errorf("Parse(%q): expected 1 message with ArrowType %v, got %+v", tt.in, tt.want, sd.Messages)
		}
	}
}

// TestCentralConnectionsRejected documents that mermaid's central-connection
// syntax (circle markers at the lifeline: "A ()->>() B", released in mermaid
// 11.16) is rejected loudly rather than silently mis-parsed. It composes with
// every arrow type, so supporting it is its own follow-up.
func TestCentralConnectionsRejected(t *testing.T) {
	for _, in := range []string{
		"sequenceDiagram\n Alice ()->>() Bob: dual",
		"sequenceDiagram\n Alice ()-x() Bob: cross dual",
		"sequenceDiagram\n Alice ()<<->>() Bob: bidirectional dual",
		"sequenceDiagram\n Alice ->>() Bob: forward",
		"sequenceDiagram\n Alice ()->> Bob: reverse",
	} {
		if _, err := Parse(in); err == nil {
			t.Errorf("expected error for central-connection syntax %q, got none", in)
		}
	}
}

// TestCrossPointBidirectionalArrows checks that mermaid's remaining message
// arrows — cross (-x/--x), async point (-)/--)) and bidirectional
// (<<->>/<<-->>) — parse to the right ArrowType.
func TestCrossPointBidirectionalArrows(t *testing.T) {
	tests := []struct {
		in   string
		want ArrowType
	}{
		{"sequenceDiagram\n A-xB: cross", SolidCross},
		{"sequenceDiagram\n A--xB: dotted cross", DottedCross},
		{"sequenceDiagram\n A-)B: async", SolidPoint},
		{"sequenceDiagram\n A--)B: dotted async", DottedPoint},
		{"sequenceDiagram\n A<<->>B: bidirectional", BidirectionalSolid},
		{"sequenceDiagram\n A<<-->>B: dotted bidirectional", BidirectionalDotted},
	}
	for _, tt := range tests {
		sd, err := Parse(tt.in)
		if err != nil {
			t.Errorf("Parse(%q): %v", tt.in, err)
			continue
		}
		if len(sd.Messages) != 1 || sd.Messages[0].ArrowType != tt.want {
			t.Errorf("Parse(%q): expected 1 message with ArrowType %v, got %+v", tt.in, tt.want, sd.Messages)
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
