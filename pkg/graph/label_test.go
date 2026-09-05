package graph

import "testing"

func TestNewGraphLabel(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		wantLines []string
		wantWidth int
	}{
		{
			name:      "single line",
			raw:       "hello",
			wantLines: []string{"hello"},
			wantWidth: 5,
		},
		{
			name:      "html line breaks",
			raw:       "line1<br/>line2<br>line3<br />line4",
			wantLines: []string{"line1", "line2", "line3", "line4"},
			wantWidth: 5,
		},
		{
			name:      "escaped newline",
			raw:       `line1\nline2`,
			wantLines: []string{"line1", "line2"},
			wantWidth: 5,
		},
		{
			name:      "display width uses runes",
			raw:       "中A",
			wantLines: []string{"中A"},
			wantWidth: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			label := newGraphLabel(tt.raw)

			if got := label.lines; len(got) != len(tt.wantLines) {
				t.Fatalf("line count = %d, want %d", len(got), len(tt.wantLines))
			}

			for i := range tt.wantLines {
				if label.lines[i] != tt.wantLines[i] {
					t.Fatalf("line %d = %q, want %q", i, label.lines[i], tt.wantLines[i])
				}
			}

			if label.width != tt.wantWidth {
				t.Fatalf("width = %d, want %d", label.width, tt.wantWidth)
			}

			if label.height() != len(tt.wantLines) {
				t.Fatalf("height = %d, want %d", label.height(), len(tt.wantLines))
			}
		})
	}
}
