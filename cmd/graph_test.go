package cmd

import (
	"testing"

	"github.com/elliotchance/orderedmap/v2"
	"github.com/stretchr/testify/assert"
)

func TestEmptyGraphHasNoNodes(t *testing.T) {
	g := mkGraph(orderedmap.NewOrderedMap[string, []textEdge]())

	assert.Equal(t, 0, len(g.nodes))
}

func TestEmptyGraphHasNoEdges(t *testing.T) {
	g := mkGraph(orderedmap.NewOrderedMap[string, []textEdge]())

	assert.Equal(t, 0, len(g.edges))
}

func TestRootNodeMappingCoords(t *testing.T) {
	data := orderedmap.NewOrderedMap[string, []textEdge]()
	data.Set("A", []textEdge{})
	data.Set("B", []textEdge{})

	g := mkGraph(data)
	g.createMapping()

	assert.Equal(t, 0, g.nodes[0].mappingCoord.x)
	assert.Equal(t, 0, g.nodes[0].mappingCoord.y)
	assert.Equal(t, 0, g.nodes[1].mappingCoord.x)
	assert.Equal(t, 1, g.nodes[1].mappingCoord.y)
}

func TestTopdownBoxes(t *testing.T) {
	mermaidMap, _, err := mermaidFileToMap(`graph TD
test --> test2`)
	if err != nil {
		t.Error(err)
	}
	drawing := drawMap(mermaidMap, nil)
	expected :=
		`+------+ 
|      | 
| test | 
|      | 
+------+ 
   |     
   |     
   v     
+-------+
|       |
| test2 |
|       |
+-------+`
	if drawing != expected {
		t.Error("Expected boxString to be ", expected, " got ", drawing)
	}
}

func TestTopdownIdenticalSizeBoxes(t *testing.T) {
	mermaidMap, _, err := mermaidFileToMap(`graph TD
test1 --> test2`)
	if err != nil {
		t.Error(err)
	}

	drawing := drawMap(mermaidMap, nil)
	expected :=
		`+-------+
|       |
| test1 |
|       |
+-------+
    |    
    |    
    v    
+-------+
|       |
| test2 |
|       |
+-------+`
	if drawing != expected {
		t.Error("Expected boxString to be ", expected, " got ", drawing)
	}
}
