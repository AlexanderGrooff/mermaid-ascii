package diagram

import (
"testing"
)

// TestNewCLIConfigWithMaxWidth verifies that maxWidth parameter is properly passed through
func TestNewCLIConfigWithMaxWidth(t *testing.T) {
tests := []struct {
name             string
maxWidth         int
expectedMaxWidth int
}{
{
name:             "unlimited_width",
maxWidth:         0,
expectedMaxWidth: 0,
},
{
name:             "width_100",
maxWidth:         100,
expectedMaxWidth: 100,
},
{
name:             "width_80",
maxWidth:         80,
expectedMaxWidth: 80,
},
{
name:             "width_200",
maxWidth:         200,
expectedMaxWidth: 200,
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
// Create config with specific maxWidth
cfg, err := NewCLIConfig(
true,  // useAscii
false, // showCoords
false, // verbose
1,     // boxBorderPadding
5,     // paddingX
5,     // paddingY
tt.maxWidth,
"LR", // graphDirection
false, // centerMultiLineLabels
)

if err != nil {
t.Fatalf("NewCLIConfig failed: %v", err)
}

if cfg.MaxWidth != tt.expectedMaxWidth {
t.Errorf("expected MaxWidth=%d, got MaxWidth=%d", tt.expectedMaxWidth, cfg.MaxWidth)
}
})
}
}

// TestNewWebConfigUsesDefaultMaxWidth verifies that NewWebConfig uses defaults
func TestNewWebConfigUsesDefaultMaxWidth(t *testing.T) {
cfg, err := NewWebConfig(
true, // useAscii
1,    // boxBorderPadding
5,    // paddingX
5,    // paddingY
)

if err != nil {
t.Fatalf("NewWebConfig failed: %v", err)
}

defaults := DefaultConfig()
if cfg.MaxWidth != defaults.MaxWidth {
t.Errorf("expected NewWebConfig to use default MaxWidth=%d, got MaxWidth=%d",
defaults.MaxWidth, cfg.MaxWidth)
}
}
