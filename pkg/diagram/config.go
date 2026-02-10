package diagram

import "fmt"

// Config holds configuration for diagram rendering.
// This replaces global variables and makes the rendering functions testable and thread-safe.
type Config struct {
	// UseAscii determines whether to use ASCII characters (true) or Unicode box-drawing characters (false)
	UseAscii bool

	// ShowCoords displays coordinate debugging information (for development)
	ShowCoords bool

	// Verbose enables detailed logging
	Verbose bool

	// --- Graph-specific configuration ---

	// BoxBorderPadding is the padding between text and border in graph nodes
	BoxBorderPadding int

	// PaddingBetweenX is the horizontal space between nodes in graphs
	PaddingBetweenX int

	// PaddingBetweenY is the vertical space between nodes in graphs
	PaddingBetweenY int

	// GraphDirection is the direction of graph layout ("LR" or "TD")
	GraphDirection string

	// StyleType determines output format for graph diagrams ("cli" or "html")
	// This controls whether graphs use colored output (html) or plain text (cli)
	StyleType string

	// --- Sequence diagram-specific configuration ---

	// SequenceParticipantSpacing is the horizontal space between participants
	SequenceParticipantSpacing int

	// SequenceMessageSpacing is the vertical space between messages (lifeline segments)
	SequenceMessageSpacing int

	// SequenceSelfMessageWidth is the width of self-message loops
	SequenceSelfMessageWidth int
}

// DefaultConfig returns a Config with sensible defaults.
// The returned config is guaranteed to pass validation.
func DefaultConfig() *Config {
	return &Config{
		UseAscii:   false, // Use Unicode by default for better appearance
		ShowCoords: false,
		Verbose:    false,
		// Graph defaults
		BoxBorderPadding: 1,
		PaddingBetweenX:  5,
		PaddingBetweenY:  5,
		GraphDirection:   "LR",
		StyleType:        "cli",
		// Sequence diagram defaults
		SequenceParticipantSpacing: 5,
		SequenceMessageSpacing:     1,
		SequenceSelfMessageWidth:   4,
	}
}

// NewConfig creates a new Config with the provided values and validates them.
// Returns an error if any values are invalid.
// For default values, use DefaultConfig() instead.
func NewConfig(useAscii bool, graphDirection, styleType string) (*Config, error) {
	config := &Config{
		UseAscii:                   useAscii,
		ShowCoords:                 false,
		Verbose:                    false,
		BoxBorderPadding:           1,
		PaddingBetweenX:            5,
		PaddingBetweenY:            5,
		GraphDirection:             graphDirection,
		StyleType:                  styleType,
		SequenceParticipantSpacing: 5,
		SequenceMessageSpacing:     1,
		SequenceSelfMessageWidth:   4,
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

func NewCLIConfig(useAscii, showCoords, verbose bool, boxBorderPadding, paddingX, paddingY int, graphDirection string) (*Config, error) {
	defaults := DefaultConfig()
	config := &Config{
		UseAscii:                   useAscii,
		ShowCoords:                 showCoords,
		Verbose:                    verbose,
		BoxBorderPadding:           boxBorderPadding,
		PaddingBetweenX:            paddingX,
		PaddingBetweenY:            paddingY,
		GraphDirection:             graphDirection,
		StyleType:                  "cli",
		SequenceParticipantSpacing: defaults.SequenceParticipantSpacing,
		SequenceMessageSpacing:     defaults.SequenceMessageSpacing,
		SequenceSelfMessageWidth:   defaults.SequenceSelfMessageWidth,
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

func NewWebConfig(useAscii bool, boxBorderPadding, paddingX, paddingY int) (*Config, error) {
	defaults := DefaultConfig()
	config := &Config{
		UseAscii:                   useAscii,
		ShowCoords:                 false,
		Verbose:                    false,
		BoxBorderPadding:           boxBorderPadding,
		PaddingBetweenX:            paddingX,
		PaddingBetweenY:            paddingY,
		GraphDirection:             "LR",
		StyleType:                  "html",
		SequenceParticipantSpacing: defaults.SequenceParticipantSpacing,
		SequenceMessageSpacing:     defaults.SequenceMessageSpacing,
		SequenceSelfMessageWidth:   defaults.SequenceSelfMessageWidth,
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// NewTestConfig creates a Config for testing with sensible defaults.
// The styleType parameter determines output format ("cli" or "html").
func NewTestConfig(useAscii bool, styleType string) *Config {
	config := DefaultConfig()
	config.UseAscii = useAscii
	config.StyleType = styleType
	return config
}

// Validate checks if the configuration values are valid.
// Returns an error if any values are invalid or would cause rendering issues.
func (c *Config) Validate() error {
	// Validate graph configuration
	if c.BoxBorderPadding < 0 {
		return &ConfigError{Field: "BoxBorderPadding", Value: c.BoxBorderPadding, Message: "must be non-negative"}
	}
	if c.PaddingBetweenX < 0 {
		return &ConfigError{Field: "PaddingBetweenX", Value: c.PaddingBetweenX, Message: "must be non-negative"}
	}
	if c.PaddingBetweenY < 0 {
		return &ConfigError{Field: "PaddingBetweenY", Value: c.PaddingBetweenY, Message: "must be non-negative"}
	}
	if c.GraphDirection != "LR" && c.GraphDirection != "TD" {
		return &ConfigError{Field: "GraphDirection", Value: c.GraphDirection, Message: "must be \"LR\" or \"TD\""}
	}
	if c.StyleType != "cli" && c.StyleType != "html" {
		return &ConfigError{Field: "StyleType", Value: c.StyleType, Message: "must be \"cli\" or \"html\""}
	}

	// Validate sequence diagram configuration
	if c.SequenceParticipantSpacing < 0 {
		return &ConfigError{Field: "SequenceParticipantSpacing", Value: c.SequenceParticipantSpacing, Message: "must be non-negative"}
	}
	if c.SequenceMessageSpacing < 0 {
		return &ConfigError{Field: "SequenceMessageSpacing", Value: c.SequenceMessageSpacing, Message: "must be non-negative"}
	}
	if c.SequenceSelfMessageWidth < 2 {
		return &ConfigError{Field: "SequenceSelfMessageWidth", Value: c.SequenceSelfMessageWidth, Message: "must be at least 2"}
	}

	return nil
}

// ConfigError represents an invalid configuration value.
type ConfigError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("invalid config: %s = %v (%s)", e.Field, e.Value, e.Message)
}
