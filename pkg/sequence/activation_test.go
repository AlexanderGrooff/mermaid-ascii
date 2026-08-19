package sequence

import (
	"strings"
	"testing"

	"github.com/AlexanderGrooff/mermaid-ascii/pkg/diagram"
)

// activeCounts returns how many activate/deactivate events each participant id
// received, so tests can assert the event stream without depending on layout.
func activeCounts(d *SequenceDiagram) (starts, ends map[string]int) {
	starts, ends = map[string]int{}, map[string]int{}
	for _, ev := range d.Events {
		switch ev.Kind {
		case EventActivate:
			starts[ev.Participant.ID]++
		case EventDeactivate:
			ends[ev.Participant.ID]++
		}
	}
	return
}

// TestActivationKeywords mirrors mermaid's 'should handle actor activation':
// the standalone keywords open and close an activation period.
func TestActivationKeywords(t *testing.T) {
	d, err := Parse("sequenceDiagram\nAlice-->>Bob:Hello Bob, how are you?\nactivate Bob\nBob-->>Alice:Hello Alice, I'm fine and you?\ndeactivate Bob")
	if err != nil {
		t.Fatal(err)
	}
	starts, ends := activeCounts(d)
	if starts["Bob"] != 1 || ends["Bob"] != 1 {
		t.Errorf("Bob activations = %d start / %d end, want 1/1", starts["Bob"], ends["Bob"])
	}
	if len(d.Messages) != 2 {
		t.Errorf("want 2 messages, got %d", len(d.Messages))
	}
}

// TestActivationShorthand mirrors 'should handle actor one line notation
// activation': "+" activates the receiver, "-" deactivates the sender.
func TestActivationShorthand(t *testing.T) {
	d, err := Parse("sequenceDiagram\nAlice-->>+Bob:Hello Bob, how are you?\nBob-->>- Alice:Hello Alice, I'm fine and you?")
	if err != nil {
		t.Fatal(err)
	}
	starts, ends := activeCounts(d)
	if starts["Bob"] != 1 || ends["Bob"] != 1 {
		t.Errorf("Bob activations = %d/%d, want 1/1", starts["Bob"], ends["Bob"])
	}
	if starts["Alice"] != 0 || ends["Alice"] != 0 {
		t.Error("+ must activate the receiver and - deactivate the sender, not Alice")
	}
	// The suffix is not part of the label or the participant name.
	if d.Messages[0].To.ID != "Bob" || d.Messages[1].To.ID != "Alice" {
		t.Errorf("suffix leaked into a name: %q, %q", d.Messages[0].To.ID, d.Messages[1].To.ID)
	}
	if got := d.Messages[0].Label; got != "Hello Bob, how are you?" {
		t.Errorf("label = %q", got)
	}
}

// TestActivationShorthandAllArrows: every arrow token accepts the suffix.
func TestActivationShorthandAllArrows(t *testing.T) {
	for _, arrow := range []string{"->>", "-->>", "->", "-->", "-x", "--x", "-)", "--)", "<<->>", "<<-->>"} {
		t.Run(arrow, func(t *testing.T) {
			d, err := Parse("sequenceDiagram\nA" + arrow + "+B: go\nB" + arrow + "-A: back")
			if err != nil {
				t.Fatalf("%s with activation suffix: %v", arrow, err)
			}
			starts, ends := activeCounts(d)
			if starts["B"] != 1 || ends["B"] != 1 {
				t.Errorf("%s: B activations = %d/%d, want 1/1", arrow, starts["B"], ends["B"])
			}
		})
	}
}

// TestStackedActivations mirrors mermaid's 'should handle stacked activations'.
func TestStackedActivations(t *testing.T) {
	d, err := Parse("sequenceDiagram\nAlice-->>+Bob:Hello Bob, how are you?\nBob-->>+Carol:Carol, let me introduce Alice?\nBob-->>- Alice:Hello Alice, please meet Carol?\nCarol->>- Bob:Oh Bob, I'm so happy to be here!")
	if err != nil {
		t.Fatal(err)
	}
	starts, ends := activeCounts(d)
	for _, id := range []string{"Bob", "Carol"} {
		if starts[id] != 1 || ends[id] != 1 {
			t.Errorf("%s activations = %d/%d, want 1/1", id, starts[id], ends[id])
		}
	}
}

// TestStackedActivationsSameParticipant: mermaid allows a participant to be
// activated twice concurrently, and requires two deactivations to balance.
func TestStackedActivationsSameParticipant(t *testing.T) {
	d, err := Parse("sequenceDiagram\nuser->>+Server: Test\nuser->>+Server: Test2\nServer->>-user: T\nServer->>-user: T2")
	if err != nil {
		t.Fatal(err)
	}
	starts, ends := activeCounts(d)
	if starts["Server"] != 2 || ends["Server"] != 2 {
		t.Errorf("Server activations = %d/%d, want 2/2", starts["Server"], ends["Server"])
	}
}

// TestActivationErrors: deactivating an inactive participant is an error, in
// both notations. The third "-" in mermaid's own unbalanced spec case fails.
func TestActivationErrors(t *testing.T) {
	cases := []struct{ name, in string }{
		{"shorthand on never-active sender", "sequenceDiagram\nA->>-B: x"},
		{"keyword on never-active participant", "sequenceDiagram\nA->>B: x\ndeactivate B"},
		{"more deactivations than activations", "sequenceDiagram\nuser->>+Server: Test\nuser->>+Server: Test2\nServer->>-user: T\nServer->>-user: T2\nServer->>-user: T3"},
		{"keyword closing a shorthand twice", "sequenceDiagram\nA->>+B: x\ndeactivate B\ndeactivate B"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := Parse(c.in)
			if err == nil || !strings.Contains(err.Error(), "inactive participant") {
				t.Errorf("want an inactive-participant error, got %v", err)
			}
		})
	}
}

// TestActivationMixedNotations: the keyword and shorthand forms share one
// counter, so one can close what the other opened.
func TestActivationMixedNotations(t *testing.T) {
	if _, err := Parse("sequenceDiagram\nA->>+B: x\ndeactivate B"); err != nil {
		t.Errorf("keyword should close a shorthand activation: %v", err)
	}
	if _, err := Parse("sequenceDiagram\nactivate B\nA->>-B: x"); err == nil {
		t.Error("- deactivates the SENDER (A), so this should still error on A")
	}
}

// TestActivationRendering: an active lifeline is drawn with the heavy stroke,
// arrows attaching to it keep their light junctions, and the period covers
// exactly the rows between activation and deactivation.
func TestActivationRendering(t *testing.T) {
	d, err := Parse("sequenceDiagram\nA->>+B: go\nB-->>-A: back\nA->>B: after")
	if err != nil {
		t.Fatal(err)
	}
	out, err := Render(d, diagram.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "┃") {
		t.Fatalf("no active lifeline drawn:\n%s", out)
	}
	if !strings.Contains(out, "┨") {
		t.Errorf("arrow leaving the active lifeline should use ┨:\n%s", out)
	}
	// The "after" message is outside the activation: its row must be light.
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "after") && strings.Contains(line, "┃") {
			t.Errorf("row after deactivation still active: %q", line)
		}
	}
}

// TestActivationOpenAtEnd: mermaid runs an unclosed activation box to the
// bottom of the diagram rather than failing, so the closing row stays active.
func TestActivationOpenAtEnd(t *testing.T) {
	d, err := Parse("sequenceDiagram\nA->>+B: x")
	if err != nil {
		t.Fatal(err)
	}
	out, err := Render(d, diagram.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if !strings.Contains(lines[len(lines)-1], "┃") {
		t.Errorf("unclosed activation should reach the last row:\n%s", out)
	}
}

// TestActivationAcrossFragment: state carries into and out of a fragment, so a
// participant activated outside stays active on the frame's inner rows.
func TestActivationAcrossFragment(t *testing.T) {
	d, err := Parse("sequenceDiagram\nA->>+B: go\nloop retry\nB->>C: sub\nend\nB-->>-A: done")
	if err != nil {
		t.Fatal(err)
	}
	out, err := Render(d, diagram.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	inLoop := false
	activeRows := 0
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "[loop retry]") {
			inLoop = true
			continue
		}
		if inLoop {
			if strings.Contains(line, "└") && strings.Contains(line, "┘") {
				break
			}
			if strings.Contains(line, "┃") || strings.Contains(line, "┠") {
				activeRows++
			}
		}
	}
	if activeRows == 0 {
		t.Errorf("B should stay active inside the fragment:\n%s", out)
	}
}

// TestDeclareAfterImplicitCreation: a participant first created by a message
// may be declared afterwards to give it a label or place it in a box, which is
// what mermaid's addActor does. Declaring the same one twice is still an error.
func TestDeclareAfterImplicitCreation(t *testing.T) {
	d, err := Parse("sequenceDiagram\nform ui->>+employees: GET\nbox Facade\nparticipant employees as Employees\nend\nemployees-->>-form ui: 200")
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Participants) != 2 {
		t.Fatalf("want 2 participants, got %d", len(d.Participants))
	}
	if d.Participants[1].Label != "Employees" {
		t.Errorf("late declaration should set the label, got %q", d.Participants[1].Label)
	}
	if len(d.Boxes) != 1 || d.Boxes[0].First != 1 || d.Boxes[0].Last != 1 {
		t.Errorf("box should claim the existing participant: %+v", d.Boxes)
	}
	if _, err := Parse("sequenceDiagram\nparticipant A\nparticipant A\nA->>A: x"); err == nil {
		t.Error("declaring the same participant twice should still error")
	}
}

// TestBoxClaimsDeclaredParticipant: when a box declares a participant that an
// earlier message already created, the box must frame THAT participant, not
// whichever one happens to have been created last.
func TestBoxClaimsDeclaredParticipant(t *testing.T) {
	d, err := Parse("sequenceDiagram\nA->>B: x\nA->>C: y\nbox Team\nparticipant B\nend\nB-->>A: z")
	if err != nil {
		t.Fatal(err)
	}
	b := d.Participants[1]
	if b.ID != "B" {
		t.Fatalf("participant order changed: %q", b.ID)
	}
	if len(d.Boxes) != 1 || d.Boxes[0].First != b.Index || d.Boxes[0].Last != b.Index {
		t.Errorf("box should frame B (index %d), got %+v", b.Index, d.Boxes[0])
	}
}

// TestActivationSpacedSuffix: mermaid's lexer skips whitespace and no name may
// begin with a sign, so the shorthand is read either side of a space.
func TestActivationSpacedSuffix(t *testing.T) {
	d, err := Parse("sequenceDiagram\nA->> +B: x\nB -->> -A: y")
	if err != nil {
		t.Fatal(err)
	}
	starts, ends := activeCounts(d)
	if starts["B"] != 1 || ends["B"] != 1 {
		t.Errorf("spaced suffix ignored: B = %d/%d, want 1/1", starts["B"], ends["B"])
	}
	if d.Messages[0].To.ID != "B" {
		t.Errorf("sign leaked into the name: %q", d.Messages[0].To.ID)
	}
}

// TestActivationUnbrokenAcrossFrames: an activation on a participant outside a
// fragment's span must stay drawn on the frame's border and divider rows, and
// on note rows, rather than flickering off for those lines.
func TestActivationUnbrokenAcrossFrames(t *testing.T) {
	cases := map[string]string{
		"fragment borders and divider": "sequenceDiagram\nA->>+B: start\nalt first\nC->>D: one\nelse second\nC->>D: two\nend\nB-->>-A: done",
		"note rows":                    "sequenceDiagram\nA->>+B: start\nNote over C,D: hello\nB-->>-A: done",
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			d, err := Parse(in)
			if err != nil {
				t.Fatal(err)
			}
			out, err := Render(d, diagram.DefaultConfig())
			if err != nil {
				t.Fatal(err)
			}
			col := -1
			for _, p := range d.Participants {
				if p.ID == "B" {
					col = p.Index
				}
			}
			// Every row between the activating and deactivating messages must
			// carry the heavy stroke at B's column.
			lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
			started := false
			for _, line := range lines {
				if strings.Contains(line, "start") {
					started = true
					continue
				}
				if strings.Contains(line, "done") {
					break
				}
				if !started || strings.TrimSpace(line) == "" {
					continue
				}
				if strings.ContainsRune(line, '►') && !strings.Contains(line, "one") && !strings.Contains(line, "two") {
					continue // the activating arrow row itself
				}
				if !strings.ContainsAny(line, "┃┠┨╂") {
					t.Errorf("activation broken on row %q (participant index %d)", line, col)
				}
			}
		})
	}
}

// TestActivationParticipantNamedActivate: a participant whose name starts with
// the keyword can still send messages (messages are matched first).
func TestActivationParticipantNamedActivate(t *testing.T) {
	d, err := Parse("sequenceDiagram\nactivate->>B: hi")
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Messages) != 1 || d.Messages[0].From.ID != "activate" {
		t.Errorf("participant named activate mishandled: %+v", d.Messages)
	}
}
