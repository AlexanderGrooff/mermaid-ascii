package sequence

import (
	"strings"
	"testing"

	"github.com/AlexanderGrooff/mermaid-ascii/pkg/diagram"
)

// altDividers returns the else-divider labels in order.
func altDividers(sd *SequenceDiagram) []string {
	var out []string
	for _, ev := range sd.Events {
		if ev.Kind == EventFragmentDivider {
			out = append(out, ev.Fragment.Label)
		}
	}
	return out
}

func hasAltStart(sd *SequenceDiagram) bool {
	for _, ev := range sd.Events {
		if ev.Kind == EventFragmentStart && ev.Fragment.Type == FragmentAlt {
			return true
		}
	}
	return false
}

// TestParseAlt covers the alt variants mermaid.js tests: a basic alt/else,
// multiple elses, a no-label alt and else, and special characters.
func TestParseAlt(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantMessages int
		wantDividers []string
	}{
		{"alt with else", "sequenceDiagram\n alt valid\n  A->>B: ok\n else invalid\n  B->>A: err\n end", 2, []string{"invalid"}},
		{"multiple elses", "sequenceDiagram\n alt a\n  A->>B: 1\n else b\n  A->>B: 2\n else c\n  A->>B: 3\n end", 3, []string{"b", "c"}},
		{"no-label alt and else", "sequenceDiagram\n alt\n  A->>B: 1\n else\n  A->>B: 2\n end", 2, []string{""}},
		{"special chars in label", "sequenceDiagram\n alt <>&! ok\n  A->>B: 1\n end", 1, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sd, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if !hasAltStart(sd) {
				t.Error("expected an alt fragment start")
			}
			if len(sd.Messages) != tt.wantMessages {
				t.Errorf("messages = %d, want %d", len(sd.Messages), tt.wantMessages)
			}
			got := altDividers(sd)
			if len(got) != len(tt.wantDividers) {
				t.Fatalf("dividers = %v, want %v", got, tt.wantDividers)
			}
			for i, w := range tt.wantDividers {
				if got[i] != w {
					t.Errorf("divider %d = %q, want %q", i, got[i], w)
				}
			}
		})
	}
}

// TestParseAltErrors covers malformed alt/else usage.
func TestParseAltErrors(t *testing.T) {
	tests := []struct {
		name, input, wantErr string
	}{
		{"else at top level", "sequenceDiagram\n A->>B: x\n else\n", "outside a matching alt"},
		{"else inside loop", "sequenceDiagram\n loop x\n  A->>B: y\n else\n end", "outside a matching alt"},
		{"unclosed alt", "sequenceDiagram\n alt x\n  A->>B: y", "unclosed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("err = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

// TestRenderAltSmoke renders an alt (incl. nested) in both charsets and checks
// the frame + divider labels appear without panicking.
func TestRenderAltSmoke(t *testing.T) {
	input := "sequenceDiagram\n alt outer\n  A->>B: x\n  alt inner\n   A->>B: y\n  else inner2\n   A->>B: z\n  end\n else outer2\n  A->>B: w\n end"
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	for _, ascii := range []bool{false, true} {
		out, err := Render(sd, diagram.NewTestConfig(ascii, "cli"))
		if err != nil {
			t.Fatalf("render ascii=%v: %v", ascii, err)
		}
		for _, want := range []string{"[alt outer]", "[outer2]", "[alt inner]", "[inner2]"} {
			if !strings.Contains(out, want) {
				t.Errorf("ascii=%v: missing %q\n%s", ascii, want, out)
			}
		}
	}
}
