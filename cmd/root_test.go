package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestResolveMaxWidth(t *testing.T) {
	var stdout *os.File
	if width, err := resolveMaxWidth("", stdout); err != nil || width != 0 {
		t.Fatalf("empty width = %d, %v", width, err)
	}
	if width, err := resolveMaxWidth("37", stdout); err != nil || width != 37 {
		t.Fatalf("numeric width = %d, %v", width, err)
	}
	for _, value := range []string{"0", "-1", "wide"} {
		if _, err := resolveMaxWidth(value, stdout); err == nil || !strings.Contains(err.Error(), "positive number or auto") {
			t.Errorf("resolveMaxWidth(%q) error = %v", value, err)
		}
	}
	if _, err := resolveMaxWidth("auto", stdout); err == nil || !strings.Contains(err.Error(), "requires stdout to be a terminal") {
		t.Errorf("non-terminal auto error = %v", err)
	}
}
