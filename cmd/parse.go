package cmd

import (
	"errors"
	"os"
	"regexp"
	"strings"

	"github.com/elliotchance/orderedmap/v2"
	log "github.com/sirupsen/logrus"
)

type labeledChild struct {
	child string
	label string
}

func setArrowWithLabel(matchedLine []string, data *orderedmap.OrderedMap[string, []labeledChild]) {
	parent := matchedLine[0]
	label := matchedLine[1]
	child := matchedLine[2]
	log.Debug("Setting arrow from ", parent, " to ", child, " with label ", label)
	setData(parent, labeledChild{child, label}, data)
}

func setArrow(matchedLine []string, data *orderedmap.OrderedMap[string, []labeledChild]) {
	parent := matchedLine[0]
	child := matchedLine[1]
	label := "" // Or should this be nil?
	log.Debug("Setting arrow from ", parent, " to ", child)
	setData(parent, labeledChild{child, label}, data)
}

func setData(parent string, child labeledChild, data *orderedmap.OrderedMap[string, []labeledChild]) {
	// Check if the parent is in the map
	if children, ok := data.Get(parent); ok {
		// If it is, append the child to the list of children
		data.Set(parent, append(children, child))
	} else {
		// If it isn't, add it to the map
		data.Set(parent, []labeledChild{child})
	}
	// Check if the child is in the map
	if _, ok := data.Get(child.child); ok {
		// If it is, do nothing
	} else {
		// If it isn't, add it to the map
		data.Set(child.child, []labeledChild{})
	}
}

func mermaidFileToMap(mermaidFile string) (*orderedmap.OrderedMap[string, []labeledChild], error) {
	// Parse the mermaid code into a map
	// The map is a tree of the form:
	// {
	//   "A": ["B", "C"],
	//   "B": ["C"],
	//   "C": []
	// }
	mermaid, err := os.ReadFile(mermaidFile)
	if err != nil {
		return nil, err
	}
	// This is an ordered map so that the output is deterministic
	// and the order of the keys is the order in which the nodes
	// are drawn
	data := orderedmap.NewOrderedMap[string, []labeledChild]()
	// Split the mermaid code into lines
	lines := strings.Split(string(mermaid), "\n")

	arrowPattern, err := regexp.Compile(`^(.+)\s+-->\s+(.+)$`)
	if err != nil {
		return nil, err
	}
	arrowWithLabelPattern, err := regexp.Compile(`^(.+)\s+-->\|(.+)\|\s+(.+)$`)
	if err != nil {
		return nil, err
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
		return nil, errors.New("first line should define the graph")
	}
	// Pop first line
	lines = lines[1:]

	// Iterate over the lines
	log.Debug("Parsing mermaid code from ", mermaidFile)
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
		} else {
			return nil, errors.New("Could not parse line: " + line)
		}
	}
	return data, nil
}
