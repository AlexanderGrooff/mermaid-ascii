package diagram

import "testing"

func TestConfigMaxWidthValidation(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxWidth = -1
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected MaxWidth validation error")
	}
}

func TestConfigFitPolicyValidation(t *testing.T) {
	cfg := DefaultConfig()
	cfg.FitPolicy = "bogus"
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected FitPolicy validation error")
	}
}
