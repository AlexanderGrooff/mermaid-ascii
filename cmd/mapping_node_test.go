package cmd

import (
	"testing"

	"github.com/elliotchance/orderedmap/v2"
	"github.com/stretchr/testify/assert"
)

func TestSingleNodeGraphHasNoEdges(t *testing.T) {
	data := orderedmap.NewOrderedMap[string, []string]()
	data.Set("A", []string{"B", "C"})
	data.Set("B", []string{"D"})
	data.Set("C", []string{"E", "F"})
	data.Set("D", []string{})
	data.Set("E", []string{})
	data.Set("F", []string{})

	g := mkGraph(data)

	assert.Equal(t, 6, len(g.nodes))
	assert.Equal(t, 5, len(g.edges))
}

func TestCreateNodeBasedOnEdge(t *testing.T) {
	data := orderedmap.NewOrderedMap[string, []string]()
	data.Set("A", []string{"B"})

	g := mkGraph(data)

	assert.Equal(t, 2, len(g.nodes))
}

func TestChildNodeHasLevelTwo(t *testing.T) {
	data := orderedmap.NewOrderedMap[string, []string]()
	data.Set("A", []string{"B"})

	g := mkGraph(data)
	g.createMapping()

	assert.Equal(t, 2, g.nodes[1].level)
}

func TestNestedChildHasLevelThree(t *testing.T) {
	data := orderedmap.NewOrderedMap[string, []string]()
	data.Set("A", []string{"B", "C"})
	data.Set("C", []string{"D"})

	g := mkGraph(data)
	g.createMapping()

	assert.Equal(t, 3, g.nodes[3].level)
}