package cmd

import (
	"testing"

	"github.com/elliotchance/orderedmap/v2"
	"github.com/google/go-cmp/cmp"
)

func TestBoxDimensions(t *testing.T) {
	boxDrawing := drawBox("text")
	if len(boxDrawing) != 8 {
		t.Error("Expected boxDrawing to have 8 columns, got ", len(boxDrawing))
	}
	if len((boxDrawing)[0]) != 5 {
		t.Error("Expected boxDrawing to have 5 rows, got ", len((boxDrawing)[0]))
	}
}

func TestDrawPlainBox(t *testing.T) {
	boxDrawing := drawBox("text")
	boxString := drawingToString(boxDrawing)
	expected :=
		`+------+
|      |
| text |
|      |
+------+`
	if boxString != expected {
		t.Error("Expected boxString to be ", expected, " got ", boxString)
	}
}

func TestDrawRightArrow(t *testing.T) {
	arrowDrawing := drawArrow(coord{0, 0}, coord{3, 0})
	boxString := drawingToString(arrowDrawing)
	expected := ` -> `
	if boxString != expected {
		t.Error("Expected boxString to be", expected, "got", boxString)
	}
}

func TestDrawStraightLowerRightArrow(t *testing.T) {
	arrowDrawing := drawArrow(coord{0, 0}, coord{3, 3})
	boxString := drawingToString(arrowDrawing)
	expected :=
		`    
\   
 \  
  > `
	if boxString != expected {
		t.Error("Expected boxString to be", expected, "got", boxString)
	}
}

func TestDrawLowerRightArrowWithLongerX(t *testing.T) {
	arrowDrawing := drawArrow(coord{0, 0}, coord{5, 3})
	boxString := drawingToString(arrowDrawing)
	expected :=
		`      
\     
 \    
  --> `
	if boxString != expected {
		t.Error("Expected boxString to be", expected, "got", boxString)
	}
}

func TestDrawLowerRightArrowWithLongerY(t *testing.T) {
	arrowDrawing := drawArrow(coord{0, 0}, coord{3, 5})
	boxString := drawingToString(arrowDrawing)
	expected :=
		`    
|   
|   
\   
 \  
  > `
	if boxString != expected {
		t.Error("Expected boxString to be", expected, "got", boxString)
	}
}

func TestDrawStraightUpperRightArrow(t *testing.T) {
	arrowDrawing := drawArrow(coord{0, 3}, coord{3, 0})
	boxString := drawingToString(arrowDrawing)
	expected :=
		`  ^ 
 /
/  `
	if boxString != expected {
		t.Error("Expected boxString to be", expected, "got", boxString)
	}
}

func TestNestedChildDrawing(t *testing.T) {
	data := orderedmap.NewOrderedMap[string, []string]()
	data.Set("A", []string{"B"})
	data.Set("B", []string{"C"})
	g := mkGraph(data)
	s := drawingToString(g.drawing)
	expected :=
		`+---+    +---+    +---+
|   |    |   |    |   |
| A |--->| B |--->| C |
|   |    |   |    |   |
+---+    +---+    +---+`
	if s != expected {
		t.Error("Expected s to be ", expected, " got ", s)
	}
}

func TestVerticalChildren(t *testing.T) {
	data := orderedmap.NewOrderedMap[string, []string]()
	data.Set("A", []string{"B", "C"})
	g := mkGraph(data)
	s := drawingToString(g.drawing)
	expected :=
		`+---+    +---+
|   |    |   |
| A |--->| B |
|   |    |   |
+---+    +---+
  \           
   \          
    \         
     \   +---+
      \  |   |
       ->| C |
         |   |
         +---+`
	if s != expected {
		t.Error("Expected s to be ", expected, " got ", s)
	}
}

func TestTopChildPointsDown(t *testing.T) {
	data := orderedmap.NewOrderedMap[string, []string]()
	data.Set("A", []string{"B", "C"})
	data.Set("B", []string{"C"})
	g := mkGraph(data)
	s := drawingToString(g.drawing)
	expected :=
		`+---+    +---+
|   |    |   |
| A |--->| B |
|   |    |   |
+---+    +---+
  \        |  
   \       |  
    \      v  
     \   +---+
      \  |   |
       ->| C |
         |   |
         +---+`
	if s != expected {
		t.Error("Expected s to be ", expected, " got ", s)
	}
}

func TestBottomChildPointsUp(t *testing.T) {
	data := orderedmap.NewOrderedMap[string, []string]()
	data.Set("A", []string{"B", "C"})
	data.Set("C", []string{"B"})
	g := mkGraph(data)
	s := drawingToString(g.drawing)
	expected :=
		`+---+    +---+
|   |    |   |
| A |--->| B |
|   |    |   |
+---+    +---+
  \        ^  
   \       |  
    \      |  
     \   +---+
      \  |   |
       ->| C |
         |   |
         +---+`
	if s != expected {
		t.Error("Expected s to be ", expected, " got ", s)
	}
}

func TestMerge(t *testing.T) {
	a := drawing{{"0", "1", "2"}}
	b := drawing{{"3", "4", "5"}}
	c := mergeDrawings(a, b, coord{0, 0})
	expected := drawing{{"3", "4", "5"}}
	if !cmp.Equal(c, expected) {
		t.Error("Expected c to be ", expected, " got ", c)
	}
}

func TestMergeWithXOffset(t *testing.T) {
	a := drawing{{"0", "1", "2"}}
	b := drawing{{"3", "4", "5"}}
	c := mergeDrawings(a, b, coord{1, 0})
	expected := drawing{{"0", "1", "2"}, {"3", "4", "5"}}
	if !cmp.Equal(c, expected) {
		t.Error("Expected c to be ", expected, " got ", c)
	}
}

func TestMergeWithYOffset(t *testing.T) {
	a := drawing{{"0", "1", "2"}}
	b := drawing{{"3", "4", "5"}}
	c := mergeDrawings(a, b, coord{0, 1})
	expected := drawing{{"0", "3", "4", "5"}}
	if !cmp.Equal(c, expected) {
		t.Error("Expected c to be ", expected, " got ", c)
	}
}
