package cmd

import (
	"fmt"

	"github.com/AlexanderGrooff/mermaid-ascii/internal/diagram"
)

func RenderDiagram(input string, config *diagram.Config) (string, error) {
	if config == nil {
		config = diagram.DefaultConfig()
	}

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

	return output, nil
}
