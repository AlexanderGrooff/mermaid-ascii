package cmd

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/elliotchance/orderedmap/v2"
	log "github.com/sirupsen/logrus"
)

type graphProperties struct {
	data           *orderedmap.OrderedMap[string, []textEdge]
	styleClasses   *map[string]styleClass
	graphDirection string
	styleType      string
	paddingX       int
	paddingY       int
	subgraphs      []*textSubgraph
	useAscii       bool
	nodeAliases    map[string]string
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

type textSubgraph struct {
	name     string
	nodes    []string
	parent   *textSubgraph
	children []*textSubgraph
}

var nodeLabelRegex = regexp.MustCompile(`^([A-Za-z0-9_.-]+)\[(.*)\]$`)

func trimOptionalQuotes(label string) string {
	label = strings.TrimSpace(label)
	if len(label) < 2 {
		return label
	}
	if (label[0] == '"' && label[len(label)-1] == '"') || (label[0] == '\'' && label[len(label)-1] == '\'') {
		return label[1 : len(label)-1]
	}
	return label
}

func parseLabeledIdentifier(raw string) (string, string, bool) {
	match := nodeLabelRegex.FindStringSubmatch(strings.TrimSpace(raw))
	if match == nil {
		return "", "", false
	}
	id := strings.TrimSpace(match[1])
	label := trimOptionalQuotes(strings.TrimSpace(match[2]))
	if label == "" {
		label = id
	}
	return id, label, true
}

func normalizeSubgraphName(name string) string {
	if _, label, ok := parseLabeledIdentifier(name); ok {
		return label
	}
	return trimOptionalQuotes(strings.TrimSpace(name))
}

func (gp *graphProperties) parseNode(line string) textNode {
	// Trim any whitespace from the line that might be left after comment removal
	trimmedLine := strings.TrimSpace(line)

	nodeWithClass, _ := regexp.Compile(`^(.+):::(.+)$`)

	if match := nodeWithClass.FindStringSubmatch(trimmedLine); match != nil {
		nodePart := strings.TrimSpace(match[1])
		styleClass := strings.TrimSpace(match[2])
		if id, label, ok := parseLabeledIdentifier(nodePart); ok {
			gp.nodeAliases[id] = label
			return textNode{label, styleClass}
		}
		if alias, ok := gp.nodeAliases[nodePart]; ok {
			return textNode{alias, styleClass}
		}
		return textNode{nodePart, styleClass}
	}
	if id, label, ok := parseLabeledIdentifier(trimmedLine); ok {
		gp.nodeAliases[id] = label
		return textNode{label, ""}
	}
	if alias, ok := gp.nodeAliases[trimmedLine]; ok {
		return textNode{alias, ""}
	}
	return textNode{trimmedLine, ""}
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
					lhs = []textNode{gp.parseNode(match[0])}
				}
				if rhs, err = gp.parseString(match[1]); err != nil {
					rhs = []textNode{gp.parseNode(match[1])}
				}
				return setArrow(lhs, rhs, gp.data), nil
			},
		},
		{
			regex: regexp.MustCompile(`^(.+)\s+-->\|(.+)\|\s+(.+)$`),
			handler: func(match []string) ([]textNode, error) {
				if lhs, err = gp.parseString(match[0]); err != nil {
					lhs = []textNode{gp.parseNode(match[0])}
				}
				if rhs, err = gp.parseString(match[2]); err != nil {
					rhs = []textNode{gp.parseNode(match[2])}
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
					node = gp.parseNode(match[0])
					lhs = []textNode{node}
				}
				if rhs, err = gp.parseString(match[1]); err != nil {
					node = gp.parseNode(match[1])
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
	rawLines := newlinePattern.Split(string(mermaid), -1)

	// Process lines to remove comments
	lines := []string{}
	for _, line := range rawLines {
		// Stop processing at "---" separator (used in test files)
		if line == "---" {
			break
		}

		// Skip lines that start with %% (comment lines)
		if strings.HasPrefix(strings.TrimSpace(line), "%%") {
			continue
		}

		// Remove inline comments (anything after %%) and trim resulting whitespace
		if idx := strings.Index(line, "%%"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}

		// Skip empty lines after comment removal
		if len(strings.TrimSpace(line)) > 0 {
			lines = append(lines, line)
		}
	}

	data := orderedmap.NewOrderedMap[string, []textEdge]()
	styleClasses := make(map[string]styleClass)
	properties := graphProperties{
		data:           data,
		styleClasses:   &styleClasses,
		graphDirection: "",
		styleType:      styleType,
		paddingX:       paddingBetweenX,
		paddingY:       paddingBetweenY,
		subgraphs:      []*textSubgraph{},
		nodeAliases:    make(map[string]string),
	}

	// Pick up optional padding directives before the graph definition
	paddingRegex := regexp.MustCompile(`^(?i)padding([xy])\s*=\s*(\d+)$`)
	for len(lines) > 0 {
		trimmed := strings.TrimSpace(lines[0])
		if trimmed == "" {
			lines = lines[1:]
			continue
		}
		if match := paddingRegex.FindStringSubmatch(trimmed); match != nil {
			paddingValue, err := strconv.Atoi(match[2])
			if err != nil {
				return &properties, err
			}
			if strings.EqualFold(match[1], "x") {
				properties.paddingX = paddingValue
			} else {
				properties.paddingY = paddingValue
			}
			lines = lines[1:]
			continue
		}
		break
	}

	if len(lines) == 0 {
		return &properties, errors.New("missing graph definition")
	}

	// First line should either say "graph TD" or "graph LR"
	switch lines[0] {
	case "graph LR", "flowchart LR":
		graphDirection = "LR"
	case "graph TD", "flowchart TD", "graph TB", "flowchart TB":
		graphDirection = "TD"
	default:
		return &properties, fmt.Errorf("unsupported graph type '%s'. Supported types: graph TD, graph TB, graph LR, flowchart TD, flowchart TB, flowchart LR", lines[0])
	}
	lines = lines[1:]

	// Track subgraph context using a stack
	subgraphStack := []*textSubgraph{}
	subgraphRegex := regexp.MustCompile(`^\s*subgraph\s+(.+)$`)
	endRegex := regexp.MustCompile(`^\s*end\s*$`)

	// Iterate over the lines
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Check for subgraph start
		if match := subgraphRegex.FindStringSubmatch(trimmedLine); match != nil {
			subgraphName := normalizeSubgraphName(match[1])
			newSubgraph := &textSubgraph{
				name:     subgraphName,
				nodes:    []string{},
				children: []*textSubgraph{},
			}

			// Set parent relationship if we're nested
			if len(subgraphStack) > 0 {
				parent := subgraphStack[len(subgraphStack)-1]
				newSubgraph.parent = parent
				parent.children = append(parent.children, newSubgraph)
			}

			subgraphStack = append(subgraphStack, newSubgraph)
			properties.subgraphs = append(properties.subgraphs, newSubgraph)
			log.Debugf("Started subgraph %s", subgraphName)
			continue
		}

		// Check for subgraph end
		if endRegex.MatchString(trimmedLine) {
			if len(subgraphStack) > 0 {
				closedSubgraph := subgraphStack[len(subgraphStack)-1]
				subgraphStack = subgraphStack[:len(subgraphStack)-1]
				log.Debugf("Ended subgraph %s", closedSubgraph.name)
			}
			continue
		}

		// Remember nodes before parsing this line
		existingNodes := make(map[string]bool)
		for el := data.Front(); el != nil; el = el.Next() {
			existingNodes[el.Key] = true
		}

		// Parse nodes and edges normally
		nodes, err := properties.parseString(line)
		if err != nil {
			log.Debugf("Parsing remaining text to node %v", line)
			node := properties.parseNode(line)
			addNode(node, properties.data)
		} else {
			// Ensure all returned nodes are in the map
			for _, node := range nodes {
				addNode(node, properties.data)
			}
		}

		// Add all new nodes to current subgraph(s)
		if len(subgraphStack) > 0 {
			for el := data.Front(); el != nil; el = el.Next() {
				nodeName := el.Key
				// If this is a new node (wasn't in existingNodes), add it to subgraph
				if !existingNodes[nodeName] {
					for _, sg := range subgraphStack {
						// Check if node is not already in the subgraph
						found := false
						for _, n := range sg.nodes {
							if n == nodeName {
								found = true
								break
							}
						}
						if !found {
							sg.nodes = append(sg.nodes, nodeName)
							log.Debugf("Added node %s to subgraph %s", nodeName, sg.name)
						}
					}
				}
			}
		}
	}
	return &properties, nil
}
