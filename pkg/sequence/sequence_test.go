package sequence

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/AlexanderGrooff/mermaid-ascii/pkg/diagram"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		wantParticipants int
		wantMessages     int
		wantErr          string
	}{
		{"empty input", "", 0, 0, "empty input"},
		{"missing sequenceDiagram keyword", "A->>B: Hello", 0, 0, "expected \"sequenceDiagram\" keyword"},
		{"only comments", "sequenceDiagram\n%% This is a comment\n%% Another comment", 0, 0, "no participants found"},
		{"no participants", "sequenceDiagram", 0, 0, "no participants found"},
		{"duplicate participant ID", "sequenceDiagram\nparticipant Alice\nparticipant Alice\nAlice->>Bob: Hi", 0, 0, "duplicate participant"},
		{"minimal diagram", "sequenceDiagram\nA->>B: Hello", 2, 1, ""},
		{"explicit participants", "sequenceDiagram\nparticipant Alice\nparticipant Bob\nAlice->>Bob: Hi", 2, 1, ""},
		{"dotted arrow", "sequenceDiagram\nA-->>B: Response", 2, 1, ""},
		{"self message", "sequenceDiagram\nA->>A: Self", 1, 1, ""},
		{"multiple messages", "sequenceDiagram\nA->>B: 1\nB->>C: 2\nC-->>A: 3", 3, 3, ""},
		{"with comments", "sequenceDiagram\n%% Comment\nA->>B: Hi %% inline comment", 2, 1, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sd, err := Parse(tt.input)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("Expected error containing %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if len(sd.Participants) != tt.wantParticipants {
				t.Errorf("Expected %d participants, got %d", tt.wantParticipants, len(sd.Participants))
			}
			if len(sd.Messages) != tt.wantMessages {
				t.Errorf("Expected %d messages, got %d", tt.wantMessages, len(sd.Messages))
			}
		})
	}
}

func TestIsSequenceDiagram(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"sequenceDiagram\nA->>B: Hello", true},
		{"SEQUENCEDIAGRAM\nA->>B: Hello", true}, // case-insensitive dispatch
		{"SequenceDiagram\nA->>B: Hello", true}, // mixed case
		{"sequenceDiagramFoo-->B", false},       // token boundary: a node id, not the keyword
		{"graph LR\nA-->B", false},
		{"graph TD\nA-->B", false},
		{"", false},
		{"%% Just a comment", false},
	}

	for _, tt := range tests {
		if got := IsSequenceDiagram(tt.input); got != tt.want {
			t.Errorf("IsSequenceDiagram(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParticipantAlias(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantID    string
		wantLabel string
	}{
		{"simple alias", "sequenceDiagram\nparticipant A as Alice\nA->>A: Hello", "A", "Alice"},
		{"no alias defaults to id", "sequenceDiagram\nparticipant Alice\nAlice->>Alice: Hi", "Alice", "Alice"},
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
			if p.ID != tt.wantID {
				t.Errorf("ID = %q, want %q", p.ID, tt.wantID)
			}
			if p.Label != tt.wantLabel {
				t.Errorf("Label = %q, want %q", p.Label, tt.wantLabel)
			}
			config := diagram.DefaultConfig()
			output, err := Render(d, config)
			if err != nil {
				t.Fatalf("render error: %v", err)
			}
			if !strings.Contains(output, tt.wantLabel) {
				t.Errorf("output should contain label %q", tt.wantLabel)
			}
		})
	}
}

func TestMessageRegex(t *testing.T) {
	tests := []struct {
		input     string
		wantFrom  string
		wantArrow string
		wantTo    string
		wantLabel string
		wantMatch bool
	}{
		{"A->>B: Hello", "A", "->>", "B", "Hello", true},
		{"A-->>B: Response", "A", "-->>", "B", "Response", true},
		{`"My Service"->>B: Test`, "My Service", "->>", "B", "Test", true},
		{"A->>B: ", "A", "->>", "B", "", true},
		{"A->B: Test", "A", "->", "B", "Test", true},
		{"A-->B: Test", "A", "-->", "B", "Test", true},
		{"A->>B", "", "", "", "", false},
		// Names with spaces, dashes and equals (mermaid ACTOR allows them).
		{"cron job->>customer-notifier: run", "cron job", "->>", "customer-notifier", "run", true},
		{"Alice-in-Wonderland->Bob: hi", "Alice-in-Wonderland", "->", "Bob", "hi", true},
		{"a=b->>c: x", "a=b", "->>", "c", "x", true},
		// A label may contain arrows and colons; only the first arrow splits.
		{"A ->> B: label with -x and <<->> and a : colon", "A", "->>", "B", "label with -x and <<->> and a : colon", true},
		// Quoted names may contain arrow characters.
		{`"A->B" ->> C: quoted`, "A->B", "->>", "C", "quoted", true},
		{`A ->> "B: with colon": label`, "A", "->>", "B: with colon", "label", true},
		// Central connections still bind to the arrow, not the name.
		{"cron job ()->>() customer-notifier: run", "cron job", "->>", "customer-notifier", "run", true},
		// Note text mentioning an arrow is not a message (name would hold ':').
		{"Note right of A: use -> to go", "", "", "", "", false},
		// Fragment openers/dividers whose label contains an arrow and a colon
		// are fragment statements, never messages (mermaid lexes the keyword
		// first). A quoted name is an explicit participant reference.
		{"else fall back -> retry: yes", "", "", "", "", false},
		{"alt cache -> hit: yes", "", "", "", "", false},
		{"loop poll -> until done: 5x", "", "", "", "", false},
		{`"loop svc" ->> B: quoted keyword name`, "loop svc", "->>", "B", "quoted keyword name", true},
	}

	for _, tt := range tests {
		parts, ok := splitMessage(tt.input)
		gotFrom, gotArrow, gotTo, gotLabel := parts.fromID, parts.arrow, parts.toID, parts.label
		if !tt.wantMatch {
			if ok {
				t.Errorf("splitMessage should not match %q", tt.input)
			}
			continue
		}
		if !ok {
			t.Fatalf("splitMessage failed to match: %q", tt.input)
		}

		if gotFrom != tt.wantFrom || gotArrow != tt.wantArrow || gotTo != tt.wantTo || gotLabel != tt.wantLabel {
			t.Errorf("splitMessage(%q) = (%q, %q, %q, %q), want (%q, %q, %q, %q)",
				tt.input, gotFrom, gotArrow, gotTo, gotLabel, tt.wantFrom, tt.wantArrow, tt.wantTo, tt.wantLabel)
		}
	}
}

func TestParticipantRegex(t *testing.T) {
	tests := []struct {
		input     string
		wantID    string
		wantLabel string
	}{
		{"participant Alice", "Alice", "Alice"},
		{"participant Alice as A", "Alice", "A"},
		{`participant "My Service"`, "My Service", "My Service"},
		{`participant "My Service" as Service`, "My Service", "Service"},
		// IDs may contain spaces, dashes and equals (mermaid's ID lexer
		// state allows them); "as" binds to its first occurrence, matching
		// mermaid for whitespace-free IDs (spaced-ID aliasing is our
		// extension — mermaid would read the whole line as one name).
		{"participant cron job", "cron job", "cron job"},
		{"participant cron job as Cron", "cron job", "Cron"},
		{"participant data=svc as DS", "data=svc", "DS"},
		{"participant a as b as c", "a", "b as c"},
	}

	for _, tt := range tests {
		sd, err := Parse("sequenceDiagram\n " + tt.input)
		if err != nil {
			t.Fatalf("Parse(%q): %v", tt.input, err)
		}
		if len(sd.Participants) != 1 {
			t.Fatalf("Parse(%q): expected 1 participant, got %d", tt.input, len(sd.Participants))
		}
		p := sd.Participants[0]
		if p.ID != tt.wantID || p.Label != tt.wantLabel {
			t.Errorf("Parse(%q) participant = (%q, %q), want (%q, %q)",
				tt.input, p.ID, p.Label, tt.wantID, tt.wantLabel)
		}
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"A->>B: Hello", []string{"A->>B: Hello"}},
		{"line1\nline2\nline3", []string{"line1", "line2", "line3"}},
		{"line1\\nline2\\nline3", []string{"line1", "line2", "line3"}},
		{"", []string{""}},
	}

	for _, tt := range tests {
		result := diagram.SplitLines(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("SplitLines(%q) len = %d, want %d", tt.input, len(result), len(tt.expected))
		}
	}
}

func TestRemoveComments(t *testing.T) {
	tests := []struct {
		input    []string
		expected []string
	}{
		{[]string{"A->>B: Hello", "B-->>A: Hi"}, []string{"A->>B: Hello", "B-->>A: Hi"}},
		{[]string{"%% This is a comment", "A->>B: Hello"}, []string{"A->>B: Hello"}},
		{[]string{"A->>B: Hello %% inline comment", "B-->>A: Hi"}, []string{"A->>B: Hello", "B-->>A: Hi"}},
	}

	for _, tt := range tests {
		result := diagram.RemoveComments(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("RemoveComments() len = %d, want %d", len(result), len(tt.expected))
		}
	}
}

func TestArrowTypeString(t *testing.T) {
	if SolidArrow.String() != "solid" {
		t.Errorf("SolidArrow.String() = %q, want \"solid\"", SolidArrow.String())
	}
	if DottedArrow.String() != "dotted" {
		t.Errorf("DottedArrow.String() = %q, want \"dotted\"", DottedArrow.String())
	}
}

func FuzzParseSequenceDiagram(f *testing.F) {
	f.Add("sequenceDiagram\nA->>B: Hello")
	f.Add("sequenceDiagram\nparticipant Alice\nAlice->>Bob: Hi")
	f.Add("sequenceDiagram\nA-->>B: Response")
	f.Add("sequenceDiagram\nA->>A: Self")

	f.Fuzz(func(t *testing.T, input string) {
		sd, err := Parse(input)
		if err != nil {
			return
		}

		for i, p := range sd.Participants {
			if p.Index != i {
				t.Errorf("Participant %q has incorrect index: got %d, expected %d", p.ID, p.Index, i)
			}
			if p.ID == "" {
				t.Errorf("Participant at index %d has empty ID", i)
			}
			if p.Label == "" {
				t.Errorf("Participant %q has empty label", p.ID)
			}
		}

		for i, msg := range sd.Messages {
			if msg.From == nil || msg.To == nil {
				t.Errorf("Message %d has nil participant", i)
			}
		}

		seen := make(map[string]bool)
		for _, p := range sd.Participants {
			if seen[p.ID] {
				t.Errorf("Duplicate participant ID: %q", p.ID)
			}
			seen[p.ID] = true
		}

		config := diagram.DefaultConfig()
		_, _ = Render(sd, config)
	})
}

func FuzzRenderSequenceDiagram(f *testing.F) {
	seeds := []string{
		"sequenceDiagram\nA->>B: Test",
		"sequenceDiagram\nA->>A: Self",
		"sequenceDiagram\nA->>B: 1\nB->>C: 2\nC->>A: 3",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		sd, err := Parse(input)
		if err != nil {
			return
		}

		for _, useAscii := range []bool{true, false} {
			config := diagram.DefaultConfig()
			config.UseAscii = useAscii

			output, err := Render(sd, config)
			if err != nil {
				return
			}

			if strings.TrimSpace(output) == "" {
				t.Error("Renderer produced empty output for valid diagram")
			}

			for _, p := range sd.Participants {
				if !strings.Contains(output, p.Label) {
					t.Errorf("Rendered output missing participant label: %q", p.Label)
				}
			}

			if !utf8.ValidString(output) {
				t.Error("Rendered output contains invalid UTF-8")
			}
		}
	})
}

func BenchmarkParse(b *testing.B) {
	tests := []struct {
		name         string
		participants int
		messages     int
	}{
		{"small_2p_5m", 2, 5},
		{"medium_5p_20m", 5, 20},
		{"large_10p_50m", 10, 50},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			input := generateDiagram(tt.participants, tt.messages)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := Parse(input)
				if err != nil {
					b.Fatalf("parse failed: %v", err)
				}
			}
		})
	}
}

func BenchmarkRender(b *testing.B) {
	tests := []struct {
		name         string
		participants int
		messages     int
	}{
		{"small_2p_5m", 2, 5},
		{"medium_5p_20m", 5, 20},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			input := generateDiagram(tt.participants, tt.messages)
			sd, err := Parse(input)
			if err != nil {
				b.Fatalf("parse failed: %v", err)
			}
			config := diagram.DefaultConfig()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, renderErr := Render(sd, config)
				if renderErr != nil {
					b.Fatalf("render error: %v", renderErr)
				}
			}
		})
	}
}

func generateDiagram(numParticipants, numMessages int) string {
	var sb strings.Builder
	sb.WriteString("sequenceDiagram\n")
	for i := 0; i < numParticipants; i++ {
		sb.WriteString("    participant P")
		sb.WriteString(string(rune('0' + i)))
		sb.WriteString("\n")
	}
	for i := 0; i < numMessages; i++ {
		from := i % numParticipants
		to := (i + 1) % numParticipants
		arrow := "-"
		if i%2 == 0 {
			arrow = "--"
		}
		sb.WriteString("    P")
		sb.WriteString(string(rune('0' + from)))
		sb.WriteString(arrow)
		sb.WriteString(">>P")
		sb.WriteString(string(rune('0' + to)))
		sb.WriteString(": Message\n")
	}
	return sb.String()
}
