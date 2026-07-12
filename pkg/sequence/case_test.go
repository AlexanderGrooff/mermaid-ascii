package sequence

import (
	"fmt"
	"testing"
)

// TestSequenceKeywordsCaseInsensitive verifies that every sequence keyword
// parses in any case. mermaid's sequence grammar declares %options
// case-insensitive, so upper/mixed-case keywords must behave identically to
// lowercase. We assert not just that it parses, but that the SEMANTICS are
// right (the keyword is classified correctly), since the risk is a captured
// keyword being switched on with the wrong case.
func TestSequenceKeywordsCaseInsensitive(t *testing.T) {
	t.Run("diagram keyword", func(t *testing.T) {
		if _, err := Parse("SEQUENCEDIAGRAM\n A->>B: x"); err != nil {
			t.Errorf("uppercase sequenceDiagram should parse: %v", err)
		}
	})

	t.Run("participant and autonumber", func(t *testing.T) {
		sd, err := Parse("sequenceDiagram\n PARTICIPANT A as Alice\n AUTONUMBER\n A->>B: x")
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		if !sd.Autonumber {
			t.Error("AUTONUMBER should enable autonumber")
		}
		if sd.Participants[0].Label != "Alice" {
			t.Errorf("participant label = %q, want Alice", sd.Participants[0].Label)
		}
	})

	t.Run("loop/end", func(t *testing.T) {
		mustFragment(t, "sequenceDiagram\n LOOP r\n  A->>B: x\n END", FragmentLoop)
	})
	t.Run("opt/end", func(t *testing.T) {
		mustFragment(t, "sequenceDiagram\n OPT r\n  A->>B: x\n END", FragmentOpt)
	})

	t.Run("alt/else/end", func(t *testing.T) {
		sd := mustFragment(t, "sequenceDiagram\n ALT a\n  A->>B: x\n ELSE b\n  A->>B: y\n END", FragmentAlt)
		if got := altDividers(sd); len(got) != 1 || got[0] != "b" {
			t.Errorf("ELSE divider = %v, want [b]", got)
		}
	})

	t.Run("note placements", func(t *testing.T) {
		mustNote(t, "sequenceDiagram\n NOTE OVER A: hi", NoteOver)
		mustNote(t, "sequenceDiagram\n Note Left Of A: hi", NoteLeftOf)
		mustNote(t, "sequenceDiagram\n note RIGHT OF A: hi", NoteRightOf)
	})

	t.Run("wrap prefix any case", func(t *testing.T) {
		n := mustNote(t, "sequenceDiagram\n Note over A:NOWRAP: hello", NoteOver)
		if n.Text != "hello" {
			t.Errorf("NOWRAP: prefix should be stripped, got %q", n.Text)
		}
	})

	t.Run("lower and upper produce identical structure", func(t *testing.T) {
		body := "%s\n participant A\n A->>B: x\n %s ok\n  B-->>A: y\n %s no\n  B-->>A: z\n %s"
		lower := fmt.Sprintf(body, "sequenceDiagram", "alt", "else", "end")
		upper := fmt.Sprintf(body, "SEQUENCEDIAGRAM", "ALT", "ELSE", "END")
		lo, err1 := Parse(lower)
		up, err2 := Parse(upper)
		if err1 != nil || err2 != nil {
			t.Fatalf("parse errors: lower=%v upper=%v", err1, err2)
		}
		if len(lo.Events) != len(up.Events) || len(lo.Messages) != len(up.Messages) {
			t.Fatalf("structure differs: lower %d events/%d msgs, upper %d/%d",
				len(lo.Events), len(lo.Messages), len(up.Events), len(up.Messages))
		}
		for i := range lo.Events {
			if lo.Events[i].Kind != up.Events[i].Kind {
				t.Errorf("event %d kind differs: %v vs %v", i, lo.Events[i].Kind, up.Events[i].Kind)
			}
		}
	})
}

func mustFragment(t *testing.T, input string, want FragmentType) *SequenceDiagram {
	t.Helper()
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("parse %q: %v", input, err)
	}
	for _, ev := range sd.Events {
		if ev.Kind == EventFragmentStart {
			if ev.Fragment.Type != want {
				t.Errorf("fragment type = %v, want %v", ev.Fragment.Type, want)
			}
			return sd
		}
	}
	t.Fatalf("no fragment start parsed from %q", input)
	return nil
}

func mustNote(t *testing.T, input string, want NotePlacement) *Note {
	t.Helper()
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("parse %q: %v", input, err)
	}
	n := firstNote(sd)
	if n == nil {
		t.Fatalf("no note parsed from %q", input)
	}
	if n.Placement != want {
		t.Errorf("placement = %v, want %v", n.Placement, want)
	}
	return n
}
