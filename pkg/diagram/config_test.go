package diagram

import "testing"

func TestConfigMaxWidthValidation(t *testing.T) {
	config := DefaultConfig()
	config.MaxWidth = 80
	if err := config.Validate(); err != nil {
		t.Fatalf("positive max width rejected: %v", err)
	}
	config.MaxWidth = -1
	if err := config.Validate(); err == nil {
		t.Fatal("negative max width accepted")
	}
}
