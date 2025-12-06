package sequence_test

import (
	"fmt"
	"log"

	"github.com/AlexanderGrooff/mermaid-ascii/internal/sequence"
)

func ExampleParse() {
	input := `sequenceDiagram
    Alice->>Bob: Hello
    Bob-->>Alice: Hi`

	sd, err := sequence.Parse(input)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Participants: %d\n", len(sd.Participants))
	fmt.Printf("Messages: %d\n", len(sd.Messages))
	// Output:
	// Participants: 2
	// Messages: 2
}

func ExampleIsSequenceDiagram() {
	sequenceInput := `sequenceDiagram
    A->>B: Test`

	graphInput := `graph LR
    A-->B`

	fmt.Printf("First is sequence: %v\n", sequence.IsSequenceDiagram(sequenceInput))
	fmt.Printf("Second is sequence: %v\n", sequence.IsSequenceDiagram(graphInput))
	// Output:
	// First is sequence: true
	// Second is sequence: false
}
