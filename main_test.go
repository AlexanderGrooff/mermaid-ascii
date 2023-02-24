package main

import (
	"testing"
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
	expected := `-->`
	if boxString != expected {
		t.Error("Expected boxString to be ", expected, " got ", boxString)
	}
}

func TestDrawDownRightArrow(t *testing.T) {
	arrowDrawing := drawArrow(coord{0, 0}, coord{3, 3})
	boxString := drawingToString(arrowDrawing)
	expected :=
		`|  
|  
|  
+->`
	if boxString != expected {
		t.Error("Expected boxString to be ", expected, " got ", boxString)
	}
}

func TestNestedChildDrawing(t *testing.T) {
	t.Skip("Has whitespace issues")
	g := mkGraph(graphData{"A": {"B"}, "B": {"C"}})
	s := drawingToString(g.drawing)
	expected :=
		`+---+     +---+     +---+
|   |     |   |     |   |
| A |---->| B |---->| C |
|   |     |   |     |   |
+---+     +---+     +---+`
	if s != expected {
		t.Error("Expected s to be ", expected, " got ", s)
	}
}
