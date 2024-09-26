package cmd

import (
	"testing"

	"github.com/elliotchance/orderedmap/v2"
	"github.com/stretchr/testify/assert"
)

func CreateChildren(parent string, children []string) []textEdge {
	edges := make([]textEdge, len(children))
	for i, child := range children {
		edges[i] = textEdge{textNode{parent, ""}, textNode{child, ""}, ""}
	}
	return edges
}

func TestSingleNodeGraphHasNoEdges(t *testing.T) {
	data := orderedmap.NewOrderedMap[string, []textEdge]()
	data.Set("A", CreateChildren("A", []string{"B", "C"}))
	data.Set("B", CreateChildren("B", []string{"D"}))
	data.Set("C", CreateChildren("C", []string{"E", "F"}))
	data.Set("D", []textEdge{})
	data.Set("E", []textEdge{})
	data.Set("F", []textEdge{})

	g := mkGraph(data)

	assert.Equal(t, 6, len(g.nodes))
	assert.Equal(t, 5, len(g.edges))
}

func TestCreateNodeBasedOnEdge(t *testing.T) {
	data := orderedmap.NewOrderedMap[string, []textEdge]()
	data.Set("A", CreateChildren("A", []string{"B"}))

	g := mkGraph(data)

	assert.Equal(t, 2, len(g.nodes))
}

func TestChildNodeMappingCoord(t *testing.T) {
	data := orderedmap.NewOrderedMap[string, []textEdge]()
	data.Set("A", CreateChildren("A", []string{"B"}))

	g := mkGraph(data)
	g.createMapping()

	assert.Equal(t, 1, g.nodes[1].gridCoord.x)
	assert.Equal(t, 0, g.nodes[1].gridCoord.y)
}

func TestNestedChildMappingCoord(t *testing.T) {
	data := orderedmap.NewOrderedMap[string, []textEdge]()
	data.Set("A", CreateChildren("A", []string{"B", "C"}))
	data.Set("C", CreateChildren("C", []string{"D"}))

	g := mkGraph(data)
	g.createMapping()

	assert.Equal(t, 2, g.nodes[3].gridCoord.x)
	assert.Equal(t, 0, g.nodes[3].gridCoord.y)
}

func TestConvertMappingToDrawingCoord(t *testing.T) {
	t.Run(
		"0,0",
		func(t *testing.T) {
			g := mkGraph(orderedmap.NewOrderedMap[string, []textEdge]())
			n := node{gridCoord: &coord{x: 0, y: 0}}

			drawCoord := *g.mappingToDrawingCoord(&n)

			assert.Equal(t, 0, drawCoord.x)
			assert.Equal(t, 0, drawCoord.y)
		},
	)
	t.Run(
		"1,0",
		func(t *testing.T) {
			g := mkGraph(orderedmap.NewOrderedMap[string, []textEdge]())
			n := node{gridCoord: &coord{x: 1, y: 0}}

			drawCoord := *g.mappingToDrawingCoord(&n)

			assert.Equal(t, 10, drawCoord.x)
			assert.Equal(t, 0, drawCoord.y)
		},
	)
	t.Run(
		"0,1",
		func(t *testing.T) {
			g := mkGraph(orderedmap.NewOrderedMap[string, []textEdge]())
			n := node{gridCoord: &coord{x: 0, y: 1}}

			drawCoord := *g.mappingToDrawingCoord(&n)

			assert.Equal(t, 0, drawCoord.x)
			assert.Equal(t, 8, drawCoord.y)
		},
	)
	t.Run(
		"1,1",
		func(t *testing.T) {
			g := mkGraph(orderedmap.NewOrderedMap[string, []textEdge]())
			n := node{gridCoord: &coord{x: 1, y: 1}}

			drawCoord := *g.mappingToDrawingCoord(&n)

			assert.Equal(t, 10, drawCoord.x)
			assert.Equal(t, 8, drawCoord.y)
		},
	)
}
