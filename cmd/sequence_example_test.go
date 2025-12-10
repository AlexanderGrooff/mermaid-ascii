package cmd

import (
	"fmt"

	"github.com/AlexanderGrooff/mermaid-ascii/internal/diagram"
	"github.com/AlexanderGrooff/mermaid-ascii/internal/sequence"
)

// ExampleParse demonstrates basic sequence diagram parsing and rendering.
func ExampleParse() {
	input := `sequenceDiagram
    Alice->>Bob: Hello
    Bob-->>Alice: Hi`

	// Parse the Mermaid syntax
	parsed, err := sequence.Parse(input)
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}

	// Render with ASCII characters
	config := diagram.NewTestConfig(true, "cli")

	output, err := sequence.Render(parsed, config)
	if err != nil {
		fmt.Printf("Render error: %v\n", err)
		return
	}

	fmt.Print(output)
	// Output:
	// +-------+     +-----+
	// | Alice |     | Bob |
	// +---+---+     +--+--+
	//     |            |
	//     | Hello      |
	//     +----------->|
	//     |            |
	//     | Hi         |
	//     |<...........+
	//     |            |
}
