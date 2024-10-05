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
	// Allow split on both \n and the actual string "\n" for curl compatibility
	newlinePattern := regexp.MustCompile(`\n|\\n`)
	lines := newlinePattern.Split(string(mermaid), -1)
	data := orderedmap.NewOrderedMap[string, []textEdge]()
	styleClasses := make(map[string]styleClass)

	patterns := map[*regexp.Regexp]func([]string){
		regexp.MustCompile(`^(.+)\s+-->\s+(.+)$`): func(match []string) {
			setArrow(match, data)
		},
		regexp.MustCompile(`^(.+)\s+-->\|(.+)\|\s+(.+)$`): func(match []string) {
			setArrowWithLabel(match, data)
		},
		regexp.MustCompile(`^classDef\s+(.+)\s+(.+)$`): func(match []string) {
			s := parseStyleClass(match)
			styleClasses[s.name] = s
		},
	}

	// First line should either say "graph TD" or "graph LR"
	switch lines[0] {
	case "graph LR", "flowchart LR":
		graphDirection = "LR"
	case "graph TD", "flowchart TD":
		graphDirection = "TD"
	default:
		return nil, nil, errors.New("first line should define the graph")
	}
	lines = lines[1:]

	// Iterate over the lines
	log.Debug("Parsing mermaid code")
	for _, line := range lines {
		if line == "" {
			continue
		}
		log.Debug("Parsing line: ", line)
		matched := false
		for pattern, handler := range patterns {
			if match := pattern.FindStringSubmatch(line); match != nil {
				handler(match[1:])
				matched = true
				break
			}
		}
		if !matched {
			return nil, nil, errors.New("Could not parse line: " + line)
		}
	}
	return data, &styleClasses, nil
}
