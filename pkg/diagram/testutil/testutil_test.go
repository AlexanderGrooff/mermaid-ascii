package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadTestCaseDecodesExpectedSpaceMarkers(t *testing.T) {
	path := filepath.Join(t.TempDir(), "case.txt")
	content := "graph LR\nA[␠input] --> B\n---\nleft␠right␠\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write test case: %v", err)
	}

	got, err := ReadTestCase(path)
	if err != nil {
		t.Fatalf("read test case: %v", err)
	}
	if got.Mermaid != "graph LR\nA[␠input] --> B\n" {
		t.Fatalf("Mermaid input changed marker: %q", got.Mermaid)
	}
	if got.Expected != "left right " {
		t.Fatalf("expected output = %q, want marker-decoded spaces", got.Expected)
	}
}
