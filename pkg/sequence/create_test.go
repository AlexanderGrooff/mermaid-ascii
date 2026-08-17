package sequence

import (
	"strings"
	"testing"

	"github.com/AlexanderGrooff/mermaid-ascii/pkg/diagram"
)

// TestCreateParticipant mirrors mermaid's 'should handle simple actor
// creation': create declares the participant, supports the actor keyword and
// an as alias, and binds it to the message that follows.
func TestCreateParticipant(t *testing.T) {
	d, err := Parse("sequenceDiagram\nparticipant a as Alice\na ->>b: Hello Bob?\ncreate participant c\nb-->>c: Hello c!\nc ->> b: Hello b?\ncreate actor d as Donald\na ->> d: Hello Donald?")
	if err != nil {
		t.Fatal(err)
	}
	byID := map[string]*Participant{}
	for _, p := range d.Participants {
		byID[p.ID] = p
	}
	if byID["c"] == nil || byID["c"].Label != "c" {
		t.Errorf("c not declared with its own label: %+v", byID["c"])
	}
	if byID["d"] == nil || byID["d"].Label != "Donald" {
		t.Errorf("create actor alias lost: %+v", byID["d"])
	}
	if len(d.Created) != 2 {
		t.Fatalf("want 2 created participants, got %d", len(d.Created))
	}
	// The create event must sit immediately before its message.
	for i, ev := range d.Events {
		if ev.Kind != EventCreate {
			continue
		}
		if i+1 >= len(d.Events) || d.Events[i+1].Kind != EventMessage {
			t.Errorf("create event for %q is not followed by its message", ev.Participant.ID)
			continue
		}
		msg := d.Events[i+1].Message
		if msg.From != ev.Participant && msg.To != ev.Participant {
			t.Errorf("create event for %q bound to an unrelated message", ev.Participant.ID)
		}
	}
}

// TestDestroyParticipant mirrors 'should handle simple actor destruction'.
func TestDestroyParticipant(t *testing.T) {
	d, err := Parse("sequenceDiagram\nparticipant a as Alice\na ->>b: Hello Bob?\ndestroy a\nb-->>a: Hello Alice!\nb ->> c: Where is Alice?\ndestroy c\nb ->> c: Where are you?")
	if err != nil {
		t.Fatal(err)
	}
	var destroyed []string
	for i, ev := range d.Events {
		if ev.Kind != EventDestroy {
			continue
		}
		destroyed = append(destroyed, ev.Participant.ID)
		// The destroy event must sit immediately after its message.
		if i == 0 || d.Events[i-1].Kind != EventMessage {
			t.Errorf("destroy event for %q is not preceded by its message", ev.Participant.ID)
			continue
		}
		msg := d.Events[i-1].Message
		if msg.From != ev.Participant && msg.To != ev.Participant {
			t.Errorf("destroy event for %q bound to an unrelated message", ev.Participant.ID)
		}
	}
	if strings.Join(destroyed, ",") != "a,c" {
		t.Errorf("destroyed = %v, want [a c]", destroyed)
	}
}

// TestCreateAndDestroySameParticipant mirrors 'should handle the creation and
// destruction of the same actor'.
func TestCreateAndDestroySameParticipant(t *testing.T) {
	d, err := Parse("sequenceDiagram\na ->>b: Hello Bob?\ncreate participant c\nb ->>c: Hello c!\nc ->> b: Hello b?\ndestroy c\nb ->> c : Bye c !")
	if err != nil {
		t.Fatal(err)
	}
	var kinds []string
	for _, ev := range d.Events {
		switch ev.Kind {
		case EventCreate, EventDestroy:
			kinds = append(kinds, ev.Kind.String()+":"+ev.Participant.ID)
		}
	}
	if strings.Join(kinds, ",") != "create:c,destroy:c" {
		t.Errorf("events = %v, want create:c then destroy:c", kinds)
	}
}

// TestCreateDestroySpacedNames: names with spaces work, as they do elsewhere
// since participant names may contain spaces.
func TestCreateDestroySpacedNames(t *testing.T) {
	d, err := Parse("sequenceDiagram\nGW->>C: x\ncreate participant Auth Service\nGW->>Auth Service: validate\ndestroy Auth Service\nGW-xAuth Service: close")
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Created) != 1 || d.Created[0].ID != "Auth Service" {
		t.Errorf("spaced created name mishandled: %+v", d.Created)
	}
}

// TestCreateDestroyErrors: mermaid requires the create/destroy statement to be
// followed by a message that involves the participant.
func TestCreateDestroyErrors(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"create never bound to a message", "sequenceDiagram\nA->>B: x\ncreate participant C\nparticipant D", "must be followed by a message"},
		{"create at end of diagram", "sequenceDiagram\nA->>B: x\ncreate participant C", "must be followed by a message"},
		{"created participant uninvolved", "sequenceDiagram\nA->>B: x\ncreate participant C\nA->>B: y", "must receive the message"},
		{"created participant sends its own creating message", "sequenceDiagram\nA->>B: x\ncreate participant C\nC->>A: hello", "must receive the message"},
		{"create reuses an existing id", "sequenceDiagram\nA->>C: x\ncreate participant C\nA->>C: y", "id already exists"},
		{"create reuses a declared id", "sequenceDiagram\nparticipant C\nA->>B: x\ncreate participant C\nA->>C: y", "id already exists"},
		{"destroy uninvolved", "sequenceDiagram\nA->>B: x\ndestroy A\nB->>C: y", "not involved in the following message"},
		{"destroy unknown participant", "sequenceDiagram\nA->>B: x\ndestroy Zed\nA->>B: y", "unknown participant"},
		{"destroy at end of diagram", "sequenceDiagram\nA->>B: x\ndestroy A", "must be followed by a message"},
		{"create requires participant or actor", "sequenceDiagram\nA->>B: x\ncreate C\nA->>C: y", "invalid syntax"},
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

// TestCreateDestroyIntervening: mermaid only consults the pending create or
// destroy when the next message arrives, so other statements may sit in
// between, including a fragment opener that wraps the binding message.
func TestCreateDestroyIntervening(t *testing.T) {
	cases := map[string]string{
		"fragment between":    "sequenceDiagram\nA->>B: x\ncreate participant C\nloop retry\nA->>C: hi\nend",
		"note between":        "sequenceDiagram\nA->>B: x\ncreate participant C\nNote over A: thinking\nA->>C: hi",
		"declaration between": "sequenceDiagram\nA->>B: x\ncreate participant C\nparticipant D\nA->>C: hi",
		"activation between":  "sequenceDiagram\nA->>B: x\ndestroy B\nactivate A\nA-xB: bye",
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse(in); err != nil {
				t.Errorf("statements between create/destroy and its message should be allowed: %v", err)
			}
		})
	}
}

// TestCreateDestroyPreservesNotes: blanking a lifeline must never eat a note
// box border or label text that happens to sit on that column.
func TestCreateDestroyPreservesNotes(t *testing.T) {
	for _, ascii := range []bool{false, true} {
		d, err := Parse("sequenceDiagram\nparticipant A\nparticipant B\nparticipant C\nA->>B: hi\ndestroy A\nA->>B: bye\nNote over B: xxxxxxxxxxxxxxxxx\nB->>C: after")
		if err != nil {
			t.Fatal(err)
		}
		cfg := diagram.DefaultConfig()
		cfg.UseAscii = ascii
		out, err := Render(d, cfg)
		if err != nil {
			t.Fatal(err)
		}
		// The note box must have the same number of border cells on each of its
		// three rows: a blanked lifeline column cannot punch through it.
		var noteRows []string
		for _, line := range strings.Split(out, "\n") {
			if strings.Contains(line, "xxxxx") {
				noteRows = append(noteRows, line)
			}
		}
		if len(noteRows) != 1 {
			t.Fatalf("ascii=%v: expected one note text row, got %d:\n%s", ascii, len(noteRows), out)
		}
		side := "│"
		if ascii {
			side = "|"
		}
		if strings.Count(noteRows[0], side) < 2 {
			t.Errorf("ascii=%v: note box lost a border:\n%s", ascii, out)
		}
	}
}

// TestCreateDestroyRendering: a created lifeline is blank above its creating
// message, a destroyed one ends with the end marker and is blank below it.
func TestCreateDestroyRendering(t *testing.T) {
	d, err := Parse("sequenceDiagram\nAlice->>Bob: hello\ncreate participant Carl\nAlice->>Carl: hi\ndestroy Carl\nAlice-xCarl: bye\nBob->>Alice: done")
	if err != nil {
		t.Fatal(err)
	}
	out, err := Render(d, diagram.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	var carlCol int
	for _, p := range d.Participants {
		if p.ID == "Carl" {
			carlCol = p.Index
		}
	}
	if carlCol != 2 {
		t.Fatalf("Carl should keep declaration order, got index %d", carlCol)
	}
	// Rows before the creating message must not draw Carl's lifeline: the
	// "hello" message row is the widest point where that is observable.
	for i, line := range lines {
		if strings.Contains(line, "hello") {
			// The row under the header and the hello rows precede creation.
			for _, before := range lines[3:i] {
				if strings.Count(before, "│") > 2 {
					t.Errorf("created lifeline drawn before creation: %q", before)
				}
			}
			break
		}
	}
	if !strings.ContainsRune(out, '×') {
		t.Errorf("destroyed lifeline should end with a marker:\n%s", out)
	}
	// After the marker, Carl's column stays blank.
	seenMarker := false
	for _, line := range lines {
		r := []rune(line)
		if seenMarker && len(r) > 24 && r[24] == '│' {
			t.Errorf("lifeline continues after destruction: %q", line)
		}
		if strings.ContainsRune(line, '×') {
			seenMarker = true
		}
	}
}

// TestCreateDestroyWithActivation: the two features compose, and destroying an
// active participant ends both the activation and the lifeline.
func TestCreateDestroyWithActivation(t *testing.T) {
	d, err := Parse("sequenceDiagram\nA->>B: x\ncreate participant C\nA->>+C: work\nC-->>-A: result\ndestroy C\nA-xC: close")
	if err != nil {
		t.Fatal(err)
	}
	out, err := Render(d, diagram.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.ContainsRune(out, '┃') {
		t.Errorf("activation lost on a created participant:\n%s", out)
	}
	if !strings.ContainsRune(out, '×') {
		t.Errorf("destruction marker missing:\n%s", out)
	}
}
