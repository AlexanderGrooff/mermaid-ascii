package cmd

import (
	"testing"
)

func TestDrawUpArrow(t *testing.T) {
	arrowDrawing := drawArrow(coord{0, 3}, coord{0, 0})
	boxString := drawingToString(arrowDrawing)
	expected := ` 
^
|
 `
	if boxString != expected {
		t.Error("Expected boxString to be", expected, "got", boxString)
	}
}

func TestDrawDownArrow(t *testing.T) {
	arrowDrawing := drawArrow(coord{0, 0}, coord{0, 3})
	boxString := drawingToString(arrowDrawing)
	expected := ` 
|
v
 `
	if boxString != expected {
		t.Error("Expected boxString to be", expected, "got", boxString)
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

func TestDrawLeftArrow(t *testing.T) {
	arrowDrawing := drawArrow(coord{3, 0}, coord{0, 0})
	boxString := drawingToString(arrowDrawing)
	expected := ` <- `
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

func TestDrawStraightLowerLeftArrow(t *testing.T) {
	arrowDrawing := drawArrow(coord{3, 0}, coord{0, 3})
	boxString := drawingToString(arrowDrawing)
	expected :=
		`    
   /
  / 
 <  `
	if boxString != expected {
		t.Error("Expected boxString to be", expected, "got", boxString)
	}
}

func TestDrawLowerLeftArrowWithLongerX(t *testing.T) {
	arrowDrawing := drawArrow(coord{5, 0}, coord{0, 3})
	boxString := drawingToString(arrowDrawing)
	expected :=
		`      
     /
    / 
 <--  `
	if boxString != expected {
		t.Error("Expected boxString to be", expected, "got", boxString)
	}
}

func TestDrawLowerLeftArrowWithLongerY(t *testing.T) {
	arrowDrawing := drawArrow(coord{3, 0}, coord{0, 5})
	boxString := drawingToString(arrowDrawing)
	expected :=
		`    
   |
   |
   /
  / 
 <  `
	if boxString != expected {
		t.Error("Expected boxString to be", expected, "got", boxString)
	}
}

func TestDrawStraightUpperRightArrow(t *testing.T) {
	arrowDrawing := drawArrow(coord{0, 3}, coord{3, 0})
	boxString := drawingToString(arrowDrawing)
	expected :=
		`  > 
 /  
/   
    `
	if boxString != expected {
		t.Error("Expected boxString to be", expected, "got", boxString)
	}
}

func TestDrawUpperRightArrowWithLongerX(t *testing.T) {
	arrowDrawing := drawArrow(coord{0, 3}, coord{5, 0})
	boxString := drawingToString(arrowDrawing)
	expected :=
		`  --> 
 /    
/     
      `
	if boxString != expected {
		t.Error("Expected boxString to be", expected, "got", boxString)
	}
}

func TestDrawUpperRightArrowWithLongerY(t *testing.T) {
	arrowDrawing := drawArrow(coord{0, 5}, coord{3, 0})
	boxString := drawingToString(arrowDrawing)
	expected :=
		`  > 
  / 
 /  
/   
|   
    `
	if boxString != expected {
		t.Error("Expected boxString to be", expected, "got", boxString)
	}
}

func TestDrawStraightUpperLeftArrow(t *testing.T) {
	arrowDrawing := drawArrow(coord{3, 3}, coord{0, 0})
	boxString := drawingToString(arrowDrawing)
	expected :=
		` <  
  \ 
   \
    `
	if boxString != expected {
		t.Error("Expected boxString to be", expected, "got", boxString)
	}
}

func TestDrawUpperLeftArrowWithLongerX(t *testing.T) {
	arrowDrawing := drawArrow(coord{5, 3}, coord{0, 0})
	boxString := drawingToString(arrowDrawing)
	expected :=
		` <--  
    \ 
     \
      `
	if boxString != expected {
		t.Error("Expected boxString to be", expected, "got", boxString)
	}
}

func TestDrawUpperLeftArrowWithVeryLongX(t *testing.T) {
	arrowDrawing := drawArrow(coord{15, 3}, coord{0, 0})
	boxString := drawingToString(arrowDrawing)
	expected :=
		` <------------  
              \ 
               \
                `
	if boxString != expected {
		t.Error("Expected boxString to be", expected, "got", boxString)
	}
}

func TestDrawUpperLeftArrowWithLongerY(t *testing.T) {
	arrowDrawing := drawArrow(coord{3, 5}, coord{0, 0})
	boxString := drawingToString(arrowDrawing)
	expected :=
		` <  
 \  
  \ 
   \
   |
    `
	if boxString != expected {
		t.Error("Expected boxString to be", expected, "got", boxString)
	}
}
