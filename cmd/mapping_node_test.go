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

func TestChildNodeMappingCoord(t *testing.T) {
	data := orderedmap.NewOrderedMap[string, []string]()
	data.Set("A", []string{"B"})

	g := mkGraph(data)

	assert.Equal(t, 1, g.nodes[1].mappingCoord.x)
	assert.Equal(t, 0, g.nodes[1].mappingCoord.y)
}

func TestNestedChildMappingCoord(t *testing.T) {
	data := orderedmap.NewOrderedMap[string, []string]()
	data.Set("A", []string{"B", "C"})
	data.Set("C", []string{"D"})

	g := mkGraph(data)

	assert.Equal(t, 2, g.nodes[3].mappingCoord.x)
	assert.Equal(t, 0, g.nodes[3].mappingCoord.y)
}

func TestConvertMappingToDrawingCoord(t *testing.T) {
	t.Run(
		"0,0",
		func(t *testing.T) {
			g := mkGraph(orderedmap.NewOrderedMap[string, []string]())
			n := node{mappingCoord: &coord{x: 0, y: 0}}

			drawCoord := g.mappingToDrawingCoord(&n)

			assert.Equal(t, 0, drawCoord.x)
			assert.Equal(t, 0, drawCoord.y)
		},
	)
	t.Run(
		"1,0",
		func(t *testing.T) {
			g := mkGraph(orderedmap.NewOrderedMap[string, []string]())
			n := node{mappingCoord: &coord{x: 1, y: 0}}

			drawCoord := g.mappingToDrawingCoord(&n)

			assert.Equal(t, 15, drawCoord.x)
			assert.Equal(t, 0, drawCoord.y)
		},
	)
	t.Run(
		"0,1",
		func(t *testing.T) {
			g := mkGraph(orderedmap.NewOrderedMap[string, []string]())
			n := node{mappingCoord: &coord{x: 0, y: 1}}

			drawCoord := g.mappingToDrawingCoord(&n)

			assert.Equal(t, 0, drawCoord.x)
			assert.Equal(t, 13, drawCoord.y)
		},
	)
	t.Run(
		"1,1",
		func(t *testing.T) {
			g := mkGraph(orderedmap.NewOrderedMap[string, []string]())
			n := node{mappingCoord: &coord{x: 1, y: 1}}

			drawCoord := g.mappingToDrawingCoord(&n)

			assert.Equal(t, 15, drawCoord.x)
			assert.Equal(t, 13, drawCoord.y)
		},
	)
}
