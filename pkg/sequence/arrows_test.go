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

// TestCentralConnections checks mermaid's central-connection syntax (circle
// markers at the lifeline, mermaid 11.16): "()" on either side of any arrow.
// The "()" is its own token in mermaid's lexer, so spaces around it are legal.
func TestCentralConnections(t *testing.T) {
	tests := []struct {
		in       string
		arrow    ArrowType
		from, to bool
	}{
		{"sequenceDiagram\n Alice ()->>() Bob: dual", SolidArrow, true, true},
		{"sequenceDiagram\n Alice ()-x() Bob: cross dual", SolidCross, true, true},
		{"sequenceDiagram\n Alice ()<<->>() Bob: bidirectional dual", BidirectionalSolid, true, true},
		{"sequenceDiagram\n Alice ->>() Bob: forward", SolidArrow, false, true},
		{"sequenceDiagram\n Alice ()->> Bob: reverse", SolidArrow, true, false},
		{"sequenceDiagram\n Alice () -->> () Bob: spaced dual", DottedArrow, true, true},
		{"sequenceDiagram\n Alice ()--)() Bob: point dual", DottedPoint, true, true},
	}
	for _, tt := range tests {
		sd, err := Parse(tt.in)
		if err != nil {
			t.Errorf("Parse(%q): %v", tt.in, err)
			continue
		}
		if len(sd.Messages) != 1 {
			t.Errorf("Parse(%q): expected 1 message, got %d", tt.in, len(sd.Messages))
			continue
		}
		m := sd.Messages[0]
		if m.ArrowType != tt.arrow || m.CentralFrom != tt.from || m.CentralTo != tt.to {
			t.Errorf("Parse(%q): got arrow=%v centralFrom=%v centralTo=%v, want %v/%v/%v",
				tt.in, m.ArrowType, m.CentralFrom, m.CentralTo, tt.arrow, tt.from, tt.to)
		}
	}
}

// TestMalformedCentralConnectionsRejected keeps near-miss circle forms loud.
func TestMalformedCentralConnectionsRejected(t *testing.T) {
	for _, in := range []string{
		"sequenceDiagram\n Alice (->>) Bob: parens around arrow",
		"sequenceDiagram\n Alice ( )->> Bob: space inside circle",
		"sequenceDiagram\n Alice (()->> Bob: unbalanced",
	} {
		if _, err := Parse(in); err == nil {
			t.Errorf("expected error for malformed central connection %q, got none", in)
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
