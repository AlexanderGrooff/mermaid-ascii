package sequence

import (
	"strings"
	"testing"

	"github.com/AlexanderGrooff/mermaid-ascii/pkg/diagram"
)

// TestActorDeclarations mirrors mermaid's actor cases (sequenceDiagram.spec.js
// apa13): `actor` declares a participant, with or without an `as` alias.
// mermaid draws actors as stick figures; ASCII renders the same labelled box.
func TestActorDeclarations(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantID    string
		wantLabel string
	}{
		{"actor with alias", "sequenceDiagram\nactor A as Alice\nactor B as Bob\nA->B: Hello Bob, how are you?\nB-->A: I am good thanks!", "A", "Alice"},
		{"actor without alias", "sequenceDiagram\nactor Alice\nAlice->>Alice: Hi", "Alice", "Alice"},
		{"uppercase AS", "sequenceDiagram\nactor A AS Database Server\nA->>A: q", "A", "Database Server"},
		{"case-insensitive keyword", "sequenceDiagram\nActor A as Alice\nA->>A: hi", "A", "Alice"},
		{"actor name with spaces", "sequenceDiagram\nactor cron job as Cron\ncron job->>cron job: tick", "cron job", "Cron"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(d.Participants) == 0 {
				t.Fatal("expected at least one participant")
			}
			p := d.Participants[0]
			if p.ID != tt.wantID || p.Label != tt.wantLabel {
				t.Errorf("got %q/%q, want %q/%q", p.ID, p.Label, tt.wantID, tt.wantLabel)
			}
			output, err := Render(d, diagram.DefaultConfig())
			if err != nil {
				t.Fatalf("render error: %v", err)
			}
			if !strings.Contains(output, tt.wantLabel) {
				t.Errorf("output should contain label %q:\n%s", tt.wantLabel, output)
			}
		})
	}
}

// TestActorParticipantMix mirrors mermaid's apa12: actors and participants
// coexist, keep declaration order, and undeclared names in messages are still
// auto-created after them.
func TestActorParticipantMix(t *testing.T) {
	d, err := Parse("sequenceDiagram\nactor Alice as Alice2\nactor Bob\nparticipant John as John2\nparticipant Mandy\nAlice->>Bob: Hi Bob\nBob->>Alice: Hi Alice\nAlice->>John: Hi John\nJohn->>Mandy: Hi Mandy\nMandy ->>Joan: Hi Joan")
	if err != nil {
		t.Fatal(err)
	}
	wantOrder := []struct{ id, label string }{
		{"Alice", "Alice2"}, {"Bob", "Bob"}, {"John", "John2"}, {"Mandy", "Mandy"}, {"Joan", "Joan"},
	}
	if len(d.Participants) != len(wantOrder) {
		t.Fatalf("want %d participants, got %d", len(wantOrder), len(d.Participants))
	}
	for i, w := range wantOrder {
		if p := d.Participants[i]; p.ID != w.id || p.Label != w.label {
			t.Errorf("participant %d = %q/%q, want %q/%q", i, p.ID, p.Label, w.id, w.label)
		}
	}
}

// TestActorDuplicateRejected: redeclaring a name (whether via actor or
// participant) is a duplicate, as for participant+participant.
func TestActorDuplicateRejected(t *testing.T) {
	if _, err := Parse("sequenceDiagram\nactor A as Alice\nparticipant A as Al\nA->>A: hi"); err == nil {
		t.Error("expected duplicate participant error")
	}
}
