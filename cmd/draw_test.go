package cmd

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestBoxDimensions(t *testing.T) {
	boxDrawing := *drawBox("text")
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

func TestMerge(t *testing.T) {
	a := &drawing{{"0", "1", "2"}}
	b := &drawing{{"3", "4", "5"}}
	c := *mergeDrawings(a, b, coord{0, 0})
	expected := drawing{{"3", "4", "5"}}
	if !cmp.Equal(c, expected) {
		t.Error("Expected c to be ", expected, " got ", c)
	}
}

func TestMergeWithXOffset(t *testing.T) {
	a := &drawing{{"0", "1", "2"}}
	b := &drawing{{"3", "4", "5"}}
	c := *mergeDrawings(a, b, coord{1, 0})
	expected := drawing{{"0", "1", "2"}, {"3", "4", "5"}}
	if !cmp.Equal(c, expected) {
		t.Error("Expected c to be ", expected, " got ", c)
	}
}

func TestMergeWithYOffset(t *testing.T) {
	a := &drawing{{"0", "1", "2"}}
	b := &drawing{{"3", "4", "5"}}
	c := *mergeDrawings(a, b, coord{0, 1})
	expected := drawing{{"0", "3", "4", "5"}}
	if !cmp.Equal(c, expected) {
		t.Error("Expected c to be ", expected, " got ", c)
	}
}
