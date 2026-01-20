package cmd

import (
	"fmt"
	"strings"

	"github.com/AlexanderGrooff/mermaid-ascii/internal/diagram"
	"github.com/AlexanderGrooff/mermaid-ascii/internal/sequence"
)

func DiagramFactory(input string) (diagram.Diagram, error) {
	input = strings.TrimSpace(input)

	if sequence.IsSequenceDiagram(input) {
		return &SequenceDiagram{}, nil
	}

	lines := strings.Split(input, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}
		if strings.HasPrefix(trimmed, "graph ") || strings.HasPrefix(trimmed, "flowchart ") {
			return &GraphDiagram{}, nil
		}
		if !strings.HasPrefix(trimmed, "%%") {
			return &GraphDiagram{}, nil
		}
	}

	return &GraphDiagram{}, nil
}

type SequenceDiagram struct {
	parsed *sequence.SequenceDiagram
}

func (sd *SequenceDiagram) Parse(input string) error {
	parsed, err := sequence.Parse(input)
	if err != nil {
		return err
	}
	sd.parsed = parsed
	return nil
}

func (sd *SequenceDiagram) Render(config *diagram.Config) (string, error) {
	if sd.parsed == nil {
		return "", fmt.Errorf("sequence diagram not parsed: call Parse() before Render()")
	}
	return sequence.Render(sd.parsed, config)
}

func (sd *SequenceDiagram) Type() string {
	return "sequence"
}

type GraphDiagram struct {
	properties *graphProperties
}

func (gd *GraphDiagram) Parse(input string) error {
	properties, err := mermaidFileToMap(input, "cli")
	if err != nil {
		return err
	}
	gd.properties = properties
	return nil
}

func (gd *GraphDiagram) Render(config *diagram.Config) (string, error) {
	if gd.properties == nil {
		return "", fmt.Errorf("graph diagram not parsed: call Parse() before Render()")
	}

	if config == nil {
		config = diagram.DefaultConfig()
	}

	styleType := config.StyleType
	if styleType == "" {
		styleType = "cli"
	}
	gd.properties.styleType = styleType
	gd.properties.useAscii = config.UseAscii
	if config.GraphDirection != "" {
		gd.properties.graphDirection = config.GraphDirection
	}
	gd.properties.paddingX = config.PaddingBetweenX
	gd.properties.paddingY = config.PaddingBetweenY
	gd.properties.boxBorderPadding = config.BoxBorderPadding

	return drawMap(gd.properties), nil
}

func (gd *GraphDiagram) Type() string {
	return "graph"
}
