package cmd

import (
	"testing"

	"github.com/elliotchance/orderedmap/v2"
	"github.com/stretchr/testify/assert"
)

func TestEmptyGraphHasNoNodes(t *testing.T) {
	g := mkGraph(orderedmap.NewOrderedMap[string, []string]())

	assert.Equal(t, 0, len(g.nodes))
}

func TestEmptyGraphHasNoEdges(t *testing.T) {
	g := mkGraph(orderedmap.NewOrderedMap[string, []string]())

	assert.Equal(t, 0, len(g.edges))
}

func TestRootNodeHasLevelOne(t *testing.T) {
	data := orderedmap.NewOrderedMap[string, []string]()
	data.Set("A", []string{})

	g := mkGraph(data)

	assert.Equal(t, 1, g.nodes[0].level)
}
