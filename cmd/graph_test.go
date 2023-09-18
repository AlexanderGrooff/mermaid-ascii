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

func TestRootNodeMappingCoords(t *testing.T) {
	data := orderedmap.NewOrderedMap[string, []string]()
	data.Set("A", []string{})
	data.Set("B", []string{})

	g := mkGraph(data)

	assert.Equal(t, 0, g.nodes[0].mappingCoord.x)
	assert.Equal(t, 0, g.nodes[0].mappingCoord.y)
	assert.Equal(t, 0, g.nodes[1].mappingCoord.x)
	assert.Equal(t, 1, g.nodes[1].mappingCoord.y)
}
