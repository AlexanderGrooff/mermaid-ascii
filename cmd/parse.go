package cmd

import (
	"errors"
	"regexp"
	"strings"

	"github.com/elliotchance/orderedmap/v2"
	log "github.com/sirupsen/logrus"
)

type graphProperties struct {
	data           *orderedmap.OrderedMap[string, []textEdge]
	styleClasses   *map[string]styleClass
	graphDirection string
	styleType      string
}

type textNode struct {
	name       string
	styleClass string
}

type textEdge struct {
	parent textNode
	child  textNode
	label  string
}

func parseNode(line string) textNode {
	nodeWithClass, _ := regexp.Compile(`^(.+):::(.+)$`)

	if match := nodeWithClass.FindStringSubmatch(line); match != nil {
		return textNode{match[1], match[2]}
	} else {
		return textNode{line, ""}
	}
}

func parseStyleClass(matchedLine []string) styleClass {
	className := matchedLine[0]
	styles := matchedLine[1]
	// Styles are comma separated and key-values are separated by colon
	// Example: fill:#f9f,stroke:#333,stroke-width:4px
	styleMap := make(map[string]string)
	for _, style := range strings.Split(styles, ",") {
		kv := strings.Split(style, ":")
		styleMap[kv[0]] = kv[1]
	}
	return styleClass{className, styleMap}
}

func setArrowWithLabel(lhs, rhs []textNode, label string, data *orderedmap.OrderedMap[string, []textEdge]) []textNode {
	log.Debug("Setting arrow from ", lhs, " to ", rhs, " with label ", label)
	for _, l := range lhs {
		for _, r := range rhs {
			setData(l, textEdge{l, r, label}, data)
		}
	}
	return rhs
}

func setArrow(lhs, rhs []textNode, data *orderedmap.OrderedMap[string, []textEdge]) []textNode {
	return setArrowWithLabel(lhs, rhs, "", data)
}

func addNode(node textNode, data *orderedmap.OrderedMap[string, []textEdge]) {
	if _, ok := data.Get(node.name); !ok {
		data.Set(node.name, []textEdge{})
	}
}

func setData(parent textNode, edge textEdge, data *orderedmap.OrderedMap[string, []textEdge]) {
	// Check if the parent is in the map
	if children, ok := data.Get(parent.name); ok {
		// If it is, append the child to the list of children
		data.Set(parent.name, append(children, edge))
	} else {
		// If it isn't, add it to the map
		data.Set(parent.name, []textEdge{edge})
	}
	// Check if the child is in the map
	if _, ok := data.Get(edge.child.name); ok {
		// If it is, do nothing
	} else {
		// If it isn't, add it to the map
		data.Set(edge.child.name, []textEdge{})
	}
}

func (gp *graphProperties) parseString(line string) ([]textNode, error) {
	log.Debugf("Parsing line: %v", line)
	var lhs, rhs []textNode
	var err error
	// Patterns are matched in order
	patterns := []struct {
		regex   *regexp.Regexp
		handler func([]string) ([]textNode, error)
	}{
		{
			regex: regexp.MustCompile(`^\s*$`),
			handler: func(match []string) ([]textNode, error) {
				// Ignore empty lines
				return []textNode{}, nil
			},
		},
		{
			regex: regexp.MustCompile(`^(.+)\s+-->\s+(.+)$`),
			handler: func(match []string) ([]textNode, error) {
				if lhs, err = gp.parseString(match[0]); err != nil {
					lhs = []textNode{parseNode(match[0])}
				}
				if rhs, err = gp.parseString(match[1]); err != nil {
					rhs = []textNode{parseNode(match[1])}
				}
				return setArrow(lhs, rhs, gp.data), nil
			},
		},
		{
			regex: regexp.MustCompile(`^(.+)\s+-->\|(.+)\|\s+(.+)$`),
			handler: func(match []string) ([]textNode, error) {
				if lhs, err = gp.parseString(match[0]); err != nil {
					lhs = []textNode{parseNode(match[0])}
				}
				if rhs, err = gp.parseString(match[2]); err != nil {
					rhs = []textNode{parseNode(match[2])}
				}
				return setArrowWithLabel(lhs, rhs, match[1], gp.data), nil
			},
		},
		{
			regex: regexp.MustCompile(`^classDef\s+(.+)\s+(.+)$`),
			handler: func(match []string) ([]textNode, error) {
				s := parseStyleClass(match)
				(*gp.styleClasses)[s.name] = s
				return []textNode{}, nil
			},
		},
		{
			regex: regexp.MustCompile(`^(.+) & (.+)$`),
			handler: func(match []string) ([]textNode, error) {
				log.Debugf("Found & pattern node %v to %v", match[0], match[1])
				var node textNode
				if lhs, err = gp.parseString(match[0]); err != nil {
					node = parseNode(match[0])
					addNode(node, gp.data)
					lhs = []textNode{node}
				}
				if rhs, err = gp.parseString(match[1]); err != nil {
					node = parseNode(match[1])
					addNode(node, gp.data)
					rhs = []textNode{node}
				}
				return append(lhs, rhs...), nil
			},
		},
	}
	for _, pattern := range patterns {
		if match := pattern.regex.FindStringSubmatch(line); match != nil {
			nodes, err := pattern.handler(match[1:])
			if err == nil {
				return nodes, nil
			}
		}
	}
	return []textNode{}, errors.New("Could not parse line: " + line)
}

func mermaidFileToMap(mermaid, styleType string) (*graphProperties, error) {
	// Allow split on both \n and the actual string "\n" for curl compatibility
	newlinePattern := regexp.MustCompile(`\n|\\n`)
	lines := newlinePattern.Split(string(mermaid), -1)
	data := orderedmap.NewOrderedMap[string, []textEdge]()
	styleClasses := make(map[string]styleClass)
	properties := graphProperties{data, &styleClasses, "", styleType}

	// First line should either say "graph TD" or "graph LR"
	switch lines[0] {
	case "graph LR", "flowchart LR":
		graphDirection = "LR"
	case "graph TD", "flowchart TD":
		graphDirection = "TD"
	default:
		return &properties, errors.New("first line should define the graph")
	}
	lines = lines[1:]

	// Iterate over the lines
	for _, line := range lines {
		_, err := properties.parseString(line)
		if err != nil {
			log.Debugf("Parsing remaining text to node %v", line)
			node := parseNode(line)
			addNode(node, properties.data)
		}
	}
	return &properties, nil
}
