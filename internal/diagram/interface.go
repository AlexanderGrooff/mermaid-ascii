package diagram

// Diagram is the interface for all diagram types (graph, sequence, etc.)
type Diagram interface {
	Parse(input string) error
	Render(config *Config) (string, error)
	Type() string
}
