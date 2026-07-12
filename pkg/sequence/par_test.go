package sequence

import (
	"strings"
	"testing"

	"github.com/AlexanderGrooff/mermaid-ascii/pkg/diagram"
)

func firstFragment(sd *SequenceDiagram) *Fragment {
	for _, ev := range sd.Events {
		if ev.Kind == EventFragmentStart {
			return ev.Fragment
		}
	}
	return nil
}

// TestParseParCriticalBreakRect covers the fragment types mermaid tests: par
// (with `and`), critical (with/without `option`), break, and rect.
func TestParseParCriticalBreakRect(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantType     FragmentType
		wantDividers []string
		wantLabel    string // opener label (after keyword)
	}{
		{"par with and", "sequenceDiagram\n par a\n  A->>B: 1\n and b\n  A->>C: 2\n end", FragmentPar, []string{"b"}, "a"},
		{"par multiple and", "sequenceDiagram\n par a\n  A->>B: 1\n and b\n  A->>B: 2\n and c\n  A->>B: 3\n end", FragmentPar, []string{"b", "c"}, "a"},
		{"critical with option", "sequenceDiagram\n critical conn\n  A->>B: 1\n option down\n  A->>B: 2\n end", FragmentCritical, []string{"down"}, "conn"},
		{"critical without option", "sequenceDiagram\n critical conn\n  A->>B: 1\n end", FragmentCritical, nil, "conn"},
		{"break", "sequenceDiagram\n break oops\n  A->>B: 1\n end", FragmentBreak, nil, "oops"},
		{"rect strips rgb", "sequenceDiagram\n rect rgb(0,255,0)\n  A->>B: 1\n end", FragmentRect, nil, ""},
		{"rect strips rgba", "sequenceDiagram\n rect rgba(0,0,0,0.1)\n  A->>B: 1\n end", FragmentRect, nil, ""},
		{"rect keeps non-color text", "sequenceDiagram\n rect highlight\n  A->>B: 1\n end", FragmentRect, nil, "highlight"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sd, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			f := firstFragment(sd)
			if f == nil {
				t.Fatal("no fragment start")
			}
			if f.Type != tt.wantType {
				t.Errorf("type = %v, want %v", f.Type, tt.wantType)
			}
			if f.Label != tt.wantLabel {
				t.Errorf("label = %q, want %q", f.Label, tt.wantLabel)
			}
			got := altDividers(sd) // reused: collects EventFragmentDivider labels
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

// TestParseDividerMismatch covers dividers used with the wrong fragment type.
func TestParseDividerMismatch(t *testing.T) {
	tests := []struct{ name, input, wantErr string }{
		{"and outside par", "sequenceDiagram\n A->>B: x\n and\n", "outside a matching par"},
		{"option outside critical", "sequenceDiagram\n A->>B: x\n option\n", "outside a matching critical"},
		{"and inside alt", "sequenceDiagram\n alt x\n  A->>B: y\n and z\n end", "outside a matching par"},
		{"option inside par", "sequenceDiagram\n par x\n  A->>B: y\n option z\n end", "outside a matching critical"},
		{"else inside par", "sequenceDiagram\n par x\n  A->>B: y\n else z\n end", "outside a matching alt"},
		{"unclosed par", "sequenceDiagram\n par x\n  A->>B: y", "unclosed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Parse(tt.input); err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("err = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

// TestRenderParCriticalBreakRectSmoke renders each in both charsets.
func TestRenderParCriticalBreakRectSmoke(t *testing.T) {
	inputs := map[string]string{
		"par":      "sequenceDiagram\n par a\n  A->>B: x\n and b\n  A->>B: y\n end",
		"critical": "sequenceDiagram\n critical c\n  A->>B: x\n option o\n  A->>B: y\n end",
		"break":    "sequenceDiagram\n break oops\n  A->>B: x\n end",
		"rect":     "sequenceDiagram\n rect rgb(0,255,0)\n  A->>B: x\n end",
	}
	for name, in := range inputs {
		for _, ascii := range []bool{false, true} {
			sd, err := Parse(in)
			if err != nil {
				t.Fatalf("%s parse: %v", name, err)
			}
			out, err := Render(sd, diagram.NewTestConfig(ascii, "cli"))
			if err != nil {
				t.Fatalf("%s render ascii=%v: %v", name, ascii, err)
			}
			if !strings.Contains(out, "["+name) {
				t.Errorf("%s ascii=%v: missing [%s ...] label:\n%s", name, ascii, name, out)
			}
		}
	}
}
