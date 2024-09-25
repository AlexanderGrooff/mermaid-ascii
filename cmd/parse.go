package cmd

import (
	"errors"
	"regexp"
	"strings"

	"github.com/elliotchance/orderedmap/v2"
	log "github.com/sirupsen/logrus"
)

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

func setArrowWithLabel(matchedLine []string, data *orderedmap.OrderedMap[string, []textEdge]) {
	parent := parseNode(matchedLine[0])
	label := matchedLine[1]
	child := parseNode(matchedLine[2])
	log.Debug("Setting arrow from ", parent, " to ", child, " with label ", label)
	setData(parent, textEdge{parent, child, label}, data)
}

func setArrow(matchedLine []string, data *orderedmap.OrderedMap[string, []textEdge]) {
	parent := parseNode(matchedLine[0])
	child := parseNode(matchedLine[1])
	label := ""
	log.Debug("Setting arrow from ", parent, " to ", child)
	setData(parent, textEdge{parent, child, label}, data)
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

func mermaidFileToMap(mermaid string) (*orderedmap.OrderedMap[string, []textEdge], *map[string]styleClass, error) {
	// Parse the mermaid code into a map
	// The map is a tree of the form:
	// {
	//   "A": ["B", "C"],
	//   "B": ["C"],
	//   "C": []
	// }

	// This is an ordered map so that the output is deterministic
	// and the order of the keys is the order in which the nodes
	// are drawn
	data := orderedmap.NewOrderedMap[string, []textEdge]()
	styleClasses := make(map[string]styleClass)
	// Split the mermaid code into lines
	lines := strings.Split(string(mermaid), "\n")

	arrowPattern, err := regexp.Compile(`^(.+)\s+-->\s+(.+)$`)
	if err != nil {
		return nil, nil, err
	}
	arrowWithLabelPattern, err := regexp.Compile(`^(.+)\s+-->\|(.+)\|\s+(.+)$`)
	if err != nil {
		return nil, nil, err
	}
	styleClassDefPattern, err := regexp.Compile(`^classDef\s+(.+)\s+(.+)$`)
	if err != nil {
		return nil, nil, err
	}

	// First line should either say "graph TD" or "graph LR"
	switch lines[0] {
	case "graph LR":
		graphDirection = "LR"
	case "graph TD":
		graphDirection = "TD"
	case "flowchart LR":
		graphDirection = "LR"
	case "flowchart TD":
		graphDirection = "TD"
	default:
		return nil, nil, errors.New("first line should define the graph")
	}
	// Pop first line
	lines = lines[1:]

	// Iterate over the lines
	log.Debug("Parsing mermaid code")
	for _, line := range lines {
		if line == "" {
			// Skip empty lines
			continue
		}
		log.Debug("Parsing line: ", line)
		if match := arrowWithLabelPattern.FindStringSubmatch(line); match != nil {
			setArrowWithLabel(match[1:], data)
		} else if match := arrowPattern.FindStringSubmatch(line); match != nil {
			setArrow(match[1:], data)
		} else if match := styleClassDefPattern.FindStringSubmatch(line); match != nil {
			s := parseStyleClass(match[1:])
			styleClasses[s.name] = s
		} else {
			return nil, nil, errors.New("Could not parse line: " + line)
		}
	}
	return data, &styleClasses, nil
}
