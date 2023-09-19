package cmd

import (
	"testing"

	"github.com/elliotchance/orderedmap/v2"
	"github.com/stretchr/testify/assert"
)

func TestSingleNodeGraphHasNoEdges(t *testing.T) {
	data := orderedmap.NewOrderedMap[string, []labeledChild]()
	data.Set("A", []labeledChild{{"B", ""}, {"C", ""}})
	data.Set("B", []labeledChild{{"D", ""}})
	data.Set("C", []labeledChild{{"E", ""}, {"F", ""}})
	data.Set("D", []labeledChild{})
	data.Set("E", []labeledChild{})
	data.Set("F", []labeledChild{})

	g := mkGraph(data)

	assert.Equal(t, 6, len(g.nodes))
	assert.Equal(t, 5, len(g.edges))
}

func TestCreateNodeBasedOnEdge(t *testing.T) {
	data := orderedmap.NewOrderedMap[string, []labeledChild]()
	data.Set("A", []labeledChild{{"B", ""}})

	g := mkGraph(data)

	assert.Equal(t, 2, len(g.nodes))
}

func TestChildNodeMappingCoord(t *testing.T) {
	data := orderedmap.NewOrderedMap[string, []labeledChild]()
	data.Set("A", []labeledChild{{"B", ""}})

	g := mkGraph(data)

	assert.Equal(t, 1, g.nodes[1].mappingCoord.x)
	assert.Equal(t, 0, g.nodes[1].mappingCoord.y)
}

func TestNestedChildMappingCoord(t *testing.T) {
	data := orderedmap.NewOrderedMap[string, []labeledChild]()
	data.Set("A", []labeledChild{{"B", ""}, {"C", ""}})
	data.Set("C", []labeledChild{{"D", ""}})

	g := mkGraph(data)

	assert.Equal(t, 2, g.nodes[3].mappingCoord.x)
	assert.Equal(t, 0, g.nodes[3].mappingCoord.y)
}

func TestConvertMappingToDrawingCoord(t *testing.T) {
	t.Run(
		"0,0",
		func(t *testing.T) {
			g := mkGraph(orderedmap.NewOrderedMap[string, []labeledChild]())
			n := node{mappingCoord: &coord{x: 0, y: 0}}

			drawCoord := *g.mappingToDrawingCoord(&n)

			assert.Equal(t, 0, drawCoord.x)
			assert.Equal(t, 0, drawCoord.y)
		},
	)
	t.Run(
		"1,0",
		func(t *testing.T) {
			g := mkGraph(orderedmap.NewOrderedMap[string, []labeledChild]())
			n := node{mappingCoord: &coord{x: 1, y: 0}}

			drawCoord := *g.mappingToDrawingCoord(&n)

			assert.Equal(t, 10, drawCoord.x)
			assert.Equal(t, 0, drawCoord.y)
		},
	)
	t.Run(
		"0,1",
		func(t *testing.T) {
			g := mkGraph(orderedmap.NewOrderedMap[string, []labeledChild]())
			n := node{mappingCoord: &coord{x: 0, y: 1}}

			drawCoord := *g.mappingToDrawingCoord(&n)

			assert.Equal(t, 0, drawCoord.x)
			assert.Equal(t, 8, drawCoord.y)
		},
	)
	t.Run(
		"1,1",
		func(t *testing.T) {
			g := mkGraph(orderedmap.NewOrderedMap[string, []labeledChild]())
			n := node{mappingCoord: &coord{x: 1, y: 1}}

			drawCoord := *g.mappingToDrawingCoord(&n)

			assert.Equal(t, 10, drawCoord.x)
			assert.Equal(t, 8, drawCoord.y)
		},
	)
}
