package cmd

import (
	"fmt"

	"github.com/AlexanderGrooff/mermaid-ascii/pkg/diagram"
)

func RenderDiagram(input string, config *diagram.Config) (string, error) {
	if config == nil {
		config = diagram.DefaultConfig()
	}

	// YAML frontmatter carries a title and theme config; the config has no
	// ASCII meaning, but the title is printed above the diagram like mermaid
	// does. Stripped here once so type detection and parsing never see it.
	input, title := diagram.StripFrontmatter(input)

	diag, err := DiagramFactory(input)
	if err != nil {
		return "", fmt.Errorf("failed to detect diagram type: %w", err)
	}

	if err := diag.Parse(input); err != nil {
		return "", fmt.Errorf("failed to parse %s diagram: %w", diag.Type(), err)
	}

	output, err := diag.Render(config)
	if err != nil {
		return "", fmt.Errorf("failed to render %s diagram: %w", diag.Type(), err)
	}

	if title != "" {
		output = title + "\n\n" + output
	}
	return output, nil
}
