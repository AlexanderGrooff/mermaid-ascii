package cmd

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

func verifyMap(t *testing.T, mermaid string, expectedMap string) {
	properties, err := mermaidFileToMap(mermaid, "cli")
	if err != nil {
		log.Fatal("Failed to parse mermaid: ", err)
	}
	actualMap := drawMap(properties)
	if expectedMap != actualMap {
		t.Errorf("Map didn't match actual map\nExpected: %v\nActual: %v", expectedMap, actualMap)
	}
}

func TestSingleNode(t *testing.T) {
	mermaid :=
		`graph LR
A`
	expectedMap :=
		`+---+
|   |
| A |
|   |
+---+`
	verifyMap(t, mermaid, expectedMap)
}

func TestSingleNodeLongerName(t *testing.T) {
	mermaid :=
		`graph LR
ABC`
	expectedMap :=
		`+-----+
|     |
| ABC |
|     |
+-----+`
	verifyMap(t, mermaid, expectedMap)
}

func TestTwoSingleRootNodes(t *testing.T) {
	mermaid :=
		`graph LR
A
B`
	expectedMap :=
		`+---+
|   |
| A |
|   |
+---+
     
     
     
     
     
+---+
|   |
| B |
|   |
+---+`
	verifyMap(t, mermaid, expectedMap)
}

func TestTwoNodesLinked(t *testing.T) {
	mermaid :=
		`graph LR
A --> B`
	expectedMap :=
		`+---+     +---+
|   |     |   |
| A |---->| B |
|   |     |   |
+---+     +---+`
	verifyMap(t, mermaid, expectedMap)
}

func TestTwoNodesLongerNames(t *testing.T) {
	mermaid :=
		`graph LR
ABC --> BCDEFG`
	expectedMap :=
		`+-----+     +--------+
|     |     |        |
| ABC |---->| BCDEFG |
|     |     |        |
+-----+     +--------+`
	verifyMap(t, mermaid, expectedMap)
}

func TestTwoLayerSingleGraph(t *testing.T) {
	mermaid :=
		`graph LR
A --> B
A --> C`
	expectedMap :=
		`+---+     +---+
|   |     |   |
| A |---->| B |
|   |     |   |
+---+     +---+
  |            
  |            
  |            
  |            
  |            
  |       +---+
  |       |   |
  +------>| C |
          |   |
          +---+`
	verifyMap(t, mermaid, expectedMap)
}

func TestBacklinkFromTop(t *testing.T) {
	mermaid :=
		`graph LR
A --> B
B --> C
A --> C
B --> D
D --> C`
	expectedMap :=
		`+---+     +---+     +---+
|   |     |   |     |   |
| A |---->| B |---->| D |
|   |     |   |     |   |
+---+     +---+     +---+
  |         |         |  
  |         |         |  
  |         |         |  
  |         |         |  
  |         v         |  
  |       +---+       |  
  |       |   |       |  
  +------>| C |<------+  
          |   |          
          +---+          `
	verifyMap(t, mermaid, expectedMap)
}

func TestBacklinkFromBottom(t *testing.T) {
	mermaid :=
		`graph LR
A --> B
B --> C
A --> C
B --> D
C --> D`
	expectedMap :=
		`+---+     +---+     +---+
|   |     |   |     |   |
| A |---->| B |---->| D |
|   |     |   |     |   |
+---+     +---+     +---+
  |         |         ^  
  |         |         |  
  |         |         |  
  |         |         |  
  |         v         |  
  |       +---+       |  
  |       |   |       |  
  +------>| C |-------+  
          |   |          
          +---+          `
	verifyMap(t, mermaid, expectedMap)
}

func TestTwoLayerSingleGraphLongerNames(t *testing.T) {
	mermaid :=
		`graph LR
ABC --> BCDEFG
ABC --> CDEFGHI`
	expectedMap :=
		`+-----+     +---------+
|     |     |         |
| ABC |---->|  BCDEFG |
|     |     |         |
+-----+     +---------+
   |                   
   |                   
   |                   
   |                   
   |                   
   |        +---------+
   |        |         |
   +------->| CDEFGHI |
            |         |
            +---------+`
	verifyMap(t, mermaid, expectedMap)
}

func TestThreeNodes(t *testing.T) {
	mermaid :=
		`graph LR
A --> B
B --> C`
	expectedMap :=
		`+---+     +---+     +---+
|   |     |   |     |   |
| A |---->| B |---->| C |
|   |     |   |     |   |
+---+     +---+     +---+`
	verifyMap(t, mermaid, expectedMap)
}

func TestThreeNodesOneLine(t *testing.T) {
	mermaid :=
		`graph LR
A --> B --> C`
	expectedMap :=
		`+---+     +---+     +---+
|   |     |   |     |   |
| A |---->| B |---->| C |
|   |     |   |     |   |
+---+     +---+     +---+`
	verifyMap(t, mermaid, expectedMap)
}

func TestTwoRootNodes(t *testing.T) {
	mermaid :=
		`graph LR
A --> B
C --> D`
	expectedMap :=
		`+---+     +---+
|   |     |   |
| A |---->| B |
|   |     |   |
+---+     +---+
               
               
               
               
               
+---+     +---+
|   |     |   |
| C |---->| D |
|   |     |   |
+---+     +---+`
	verifyMap(t, mermaid, expectedMap)
}

func TestTwoRootNodesLongerNames(t *testing.T) {
	mermaid :=
		`graph LR
ABC --> BCDEFG
CDEFGH --> DEF`
	expectedMap :=
		`+--------+     +--------+
|        |     |        |
|  ABC   |---->| BCDEFG |
|        |     |        |
+--------+     +--------+
                         
                         
                         
                         
                         
+--------+     +--------+
|        |     |        |
| CDEFGH |---->|  DEF   |
|        |     |        |
+--------+     +--------+`
	verifyMap(t, mermaid, expectedMap)
}

func TestAmpersandLHS(t *testing.T) {
	mermaid :=
		`graph LR
A & B --> C`
	expectedMap :=
		`+---+     +---+
|   |     |   |
| A |---->| C |
|   |     |   |
+---+     +---+
            ^  
            |  
            |  
            |  
            |  
+---+       |  
|   |       |  
| B |-------+  
|   |          
+---+          `
	verifyMap(t, mermaid, expectedMap)
}

func TestAmpersandRHS(t *testing.T) {
	mermaid :=
		`graph LR
A --> B & C`
	expectedMap :=
		`+---+     +---+
|   |     |   |
| A |---->| B |
|   |     |   |
+---+     +---+
  |            
  |            
  |            
  |            
  |            
  |       +---+
  |       |   |
  +------>| C |
          |   |
          +---+`
	verifyMap(t, mermaid, expectedMap)
}

func TestAmpersandLHSAndRHS(t *testing.T) {
	mermaid :=
		`graph LR
A & B --> C & D`
	expectedMap :=
		`+---+     +---+
|   |     |   |
| A |--+->| C |
|   |  |  |   |
+---+  |  +---+
  |    |       
  |    |       
  +----+       
  |    |       
  |    |       
+---+  |  +---+
|   |  |  |   |
| B |--+->| D |
|   |     |   |
+---+     +---+`
	verifyMap(t, mermaid, expectedMap)
}

func TestAmpersandWithoutEdge(t *testing.T) {
	mermaid :=
		`graph LR
A & B`
	expectedMap :=
		`+---+
|   |
| A |
|   |
+---+
     
     
     
     
     
+---+
|   |
| B |
|   |
+---+`
	verifyMap(t, mermaid, expectedMap)
}

func TestSelfReference(t *testing.T) {
	mermaid :=
		`graph LR
A --> A`
	expectedMap :=
		`+---+  
|   |  
| A |-+
|   | |
+---+ |
  ^   |
  +---+`
	verifyMap(t, mermaid, expectedMap)
}

func TestSelfReferenceWithEdge(t *testing.T) {
	mermaid :=
		`graph LR
A --> A & B`
	expectedMap :=
		`+---+     +---+
|   |     |   |
| A |--+->| B |
|   |  |  |   |
+---+  |  +---+
  ^    |       
  +----+       `
	verifyMap(t, mermaid, expectedMap)
}

func TestBackReferenceFromChild(t *testing.T) {
	mermaid :=
		`graph LR
A --> B --> C --> A`
	expectedMap :=
		`+---+     +---+     +---+
|   |     |   |     |   |
| A |<-+->| B |--+->| C |
|   |  |  |   |  |  |   |
+---+  |  +---+  |  +---+
       |         |       
       +---------+       `
	verifyMap(t, mermaid, expectedMap)
}

func TestPreserveOrderOfDefinition(t *testing.T) {
	mermaid :=
		`graph LR
A
B
B --> A
A --> A
B --> C
C --> A`
	expectedMap :=
		`+---+     +---+
|   |     |   |
| A |<-+--| C |
|   |  |  |   |
+---+  |  +---+
  ^    |    ^  
  |    |    |  
  +----+    |  
  |         |  
  |         |  
+---+       |  
|   |       |  
| B |-------+  
|   |          
+---+          `
	verifyMap(t, mermaid, expectedMap)
}
