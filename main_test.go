package main

import (
	"testing"
)

func TestBoxDimensions(t *testing.T) {
	boxDrawing := drawBox("text")
	if len(*boxDrawing) != 8 {
		t.Error("Expected boxDrawing to have 8 columns, got ", len(*boxDrawing))
	}
	if len((*boxDrawing)[0]) != 5 {
		t.Error("Expected boxDrawing to have 5 rows, got ", len((*boxDrawing)[0]))
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
