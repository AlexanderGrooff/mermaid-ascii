package cmd

import (
	"errors"
	"os"
	"strings"

	"github.com/elliotchance/orderedmap/v2"
	log "github.com/sirupsen/logrus"
)

func mermaidFileToMap(mermaidFile string) (*orderedmap.OrderedMap[string, []string], error) {
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
	data := orderedmap.NewOrderedMap[string, []string]()
	// Split the mermaid code into lines
	lines := strings.Split(string(mermaid), "\n")
	// Iterate over the lines
	log.Debug("Parsing mermaid code from ", mermaidFile)
	for _, line := range lines {
		if line == "" {
			// Skip empty lines
			continue
		}
		log.Debug("Parsing line: ", line)
		// Split the line into words
		words := strings.Split(line, " ")
		// The first word is the parent
		parent := words[0]
		// The second word is the arrow
		arrow := words[1]
		// The third word is the child
		child := words[2]
		// Check that the arrow is "-->"
		if arrow != "-->" {
			return nil, errors.New("Invalid arrow")
		}
		// Check if the parent is in the map
		if children, ok := data.Get(parent); ok {
			// If it is, append the child to the list of children
			data.Set(parent, append(children, child))
		} else {
			// If it isn't, add it to the map
			data.Set(parent, []string{child})
		}
		// Check if the child is in the map
		if _, ok := data.Get(child); ok {
			// If it is, do nothing
		} else {
			// If it isn't, add it to the map
			data.Set(child, []string{})
		}
	}
	return data, nil
}
