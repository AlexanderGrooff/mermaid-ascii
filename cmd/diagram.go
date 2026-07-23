package cmd

import (
	"fmt"
	"strings"

	"github.com/AlexanderGrooff/mermaid-ascii/pkg/diagram"
	"github.com/AlexanderGrooff/mermaid-ascii/pkg/er"
	"github.com/AlexanderGrooff/mermaid-ascii/pkg/sequence"
)

func DiagramFactory(input string) (diagram.Diagram, error) {
	input = strings.TrimSpace(input)

	if sequence.IsSequenceDiagram(input) {
		return &SequenceDiagram{}, nil
	}

	if er.IsErDiagram(input) {
		return &ErDiagram{}, nil
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
	gd.properties.boxBorderPadding = config.BoxBorderPadding
	gd.properties.paddingX = config.PaddingBetweenX
	gd.properties.paddingY = config.PaddingBetweenY
	gd.properties.styleType = styleType
	gd.properties.useAscii = config.UseAscii

	return drawMap(gd.properties), nil
}

func (gd *GraphDiagram) Type() string {
	return "graph"
}

// ErDiagram adapts the er package to the Diagram interface.
type ErDiagram struct {
	parsed *er.ErDiagram
}

func (d *ErDiagram) Parse(input string) error {
	parsed, err := er.Parse(input)
	if err != nil {
		return err
	}
	d.parsed = parsed
	return nil
}

func (d *ErDiagram) Render(config *diagram.Config) (string, error) {
	if d.parsed == nil {
		return "", fmt.Errorf("er diagram not parsed: call Parse() before Render()")
	}
	return er.Render(d.parsed, config.UseAscii), nil
}

func (d *ErDiagram) Type() string { return "er" }
