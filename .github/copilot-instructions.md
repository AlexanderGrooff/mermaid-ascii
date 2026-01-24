# GitHub Copilot Instructions for mermaid-ascii

# 🛑🛑🛑 STOP - MANDATORY PRE-FLIGHT - READ THIS BEFORE RESPONDING 🛑🛑🛑

☐ State which user and project copilot-instructions.md sections apply to this request
☐ Check if any Agent Skills apply (list them explicitly)
☐ If multi-step work: Create todo list with #manage_todo_list
☐ Mark tasks in-progress and completed as you work
☐ Use #code-review before finalizing ANY code changes
☐ Monitor and report token usage at checkpoints (700K/850K/950K)

**If you cannot check ALL boxes above, STOP and ask for clarification.**

**Example Response Format:**
```
**Following copilot-instructions.md sections: Go Testing, Build Patterns**
**Applicable Agent Skills: #go-testing, #code-review**
**Will use #manage_todo_list for multi-step tracking**

We need to...
```

---

# 📖 REQUIRED READING

**ALWAYS read the user-level copilot-instructions.md file first:**
- **Location**: `/home/warnes/src/vscode-config/copilot-instructions.md`
- **Contains**: Communication style, token monitoring, cross-project development patterns
- **Why**: Establishes baseline behavior and standards across all projects

**This file (project-specific) provides:**
- Go development best practices and anti-patterns
- Testing patterns with table-driven tests
- Build and release workflows
- mermaid-ascii specific patterns (diagram rendering, text wrapping, CLI flags)
- Agent Skills specific to Go development

---

## Quick Skill Reference

**Workflow & Quality:**
- **#code-review** - REQUIRED before finalizing any code changes
- **#git-commit-message** - For commit message generation  
- **#manage_todo_list** - For multi-step task tracking and planning

**Go Development:**
- **#go-testing** - Table-driven tests, test coverage, Go testing patterns
- **#go-build-and-test** - Build, test, and release workflows
- **#go-struct-patterns** - Struct initialization, configuration patterns

---

## Project Overview

**mermaid-ascii** is a Go-based tool that converts Mermaid diagram syntax to ASCII art diagrams for terminal/documentation display.

**This fork adds:**
- `<br/>` and `<br>` HTML tag support for multi-line node labels
- `-w/--maxWidth` CLI flag for diagram width control
- Enhanced text wrapping with proper line splitting

**Core Components:**
- `cmd/` - CLI commands, parsing, rendering, graph layout
- `internal/diagram/` - Configuration, validation, rendering infrastructure
- `internal/sequence/` - Sequence diagram specific rendering
- `main.go` - Entry point, Cobra CLI setup

---

## ⚠️ CRITICAL WORKFLOW CHECKLIST

**Before implementing ANY code changes, verify you will:**

1. ✅ **Create/update unit tests** - Go requires tests alongside code
2. ✅ **Follow anti-patterns** - Check relevant sections below before coding
3. ✅ **Review changes** - Use systematic code review before finalizing
4. ✅ **Update documentation** - Update comments and README for exported functions

**After making changes, verify you have:**

1. ✅ **Tests passing** - All new/modified code has passing tests
2. ✅ **Documentation updated** - Comments and README current
3. ✅ **No anti-patterns** - Reviewed against project-specific warnings
4. ✅ **User informed** - Confirmed completion to user

---

## Go Development Best Practices

### CRITICAL: Always Write Tests

**Go convention: Tests live alongside code in `*_test.go` files**

```go
// ✅ CORRECT - Test file next to implementation
cmd/
  graph.go
  graph_test.go          // Tests for graph.go
  parse.go
  parse_test.go          // Tests for parse.go
```

**NEVER commit code without tests:**
```go
// ❌ INCORRECT - No test file
cmd/
  new_feature.go  // No new_feature_test.go!

// ✅ CORRECT - Test file included
cmd/
  new_feature.go
  new_feature_test.go
```

### Table-Driven Tests

**ALWAYS use table-driven test pattern for multiple scenarios:**

```go
// ✅ CORRECT - Table-driven test
func TestFeature(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "case1", input: "a", expected: "A"},
		{name: "case2", input: "b", expected: "B"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Feature(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// ❌ INCORRECT - Separate test functions (harder to maintain)
func TestFeatureCase1(t *testing.T) { /* ... */ }
func TestFeatureCase2(t *testing.T) { /* ... */ }
```

### Error Handling

**Always check and handle errors explicitly:**

```go
// ✅ CORRECT - Explicit error handling
result, err := SomeFunction()
if err != nil {
	return fmt.Errorf("failed to do something: %w", err)
}

// ❌ INCORRECT - Ignoring errors
result, _ := SomeFunction()  // Silent failure!
```

### Struct Initialization

**Use named fields for clarity:**

```go
// ✅ CORRECT - Named fields
config := &Config{
	MaxWidth:     100,
	PaddingX:     5,
	PaddingY:     2,
	UseAscii:     true,
}

// ❌ INCORRECT - Positional (brittle if struct changes)
config := &Config{100, 5, 2, true}
```

---

## mermaid-ascii Specific Patterns

### Text Processing

**String manipulation best practices:**

```go
// ✅ CORRECT - Use strings package for efficiency
name := strings.ReplaceAll(strings.ReplaceAll(n.name, "<br/>", "\n"), "<br>", "\n")

// ✅ CORRECT - Use strings.Builder for concatenation in loops
var sb strings.Builder
for _, line := range lines {
	sb.WriteString(line)
	sb.WriteString("\n")
}
result := sb.String()

// ❌ INCORRECT - Repeated string concatenation (inefficient)
result := ""
for _, line := range lines {
	result += line + "\n"  // Creates new string each iteration!
}
```

### Configuration Management

**Config structs should validate themselves:**

```go
// ✅ CORRECT - Validation method
type Config struct {
	MaxWidth int
	PaddingX int
}

func (c *Config) Validate() error {
	if c.MaxWidth < 0 {
		return fmt.Errorf("maxWidth cannot be negative")
	}
	if c.PaddingX < 0 {
		return fmt.Errorf("paddingX cannot be negative")
	}
	return nil
}

// Usage
config := NewConfig(...)
if err := config.Validate(); err != nil {
	return err
}
```

### CLI Flag Patterns (Cobra)

**Use persistent flags for global options:**

```go
// ✅ CORRECT - Persistent flags available to all subcommands
var verbose bool
rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

// ✅ CORRECT - IntVarP for integer flags with short form
var maxWidth int
rootCmd.PersistentFlags().IntVarP(&maxWidth, "maxWidth", "w", 0, "Maximum width")
```

---

## Testing Patterns

### Test File Organization

```go
package cmd

import (
	"testing"
	"strings"
	"github.com/AlexanderGrooff/mermaid-ascii/internal/diagram"
)

// Helper functions at top
func helperFunction(t *testing.T, input string) string {
	t.Helper()  // Mark as helper for better error reporting
	// ... helper logic
}

// Table-driven tests
func TestMainFeature(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Test cases
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test logic
		})
	}
}

// Edge cases in separate tests if needed
func TestEdgeCase(t *testing.T) {
	// Specific edge case
}
```

### Test Output Validation

```go
// ✅ CORRECT - Clear error messages with actual output
if !strings.Contains(output, expected) {
	t.Errorf("expected output to contain %q, got:\n%s", expected, output)
}

// ✅ CORRECT - Helper functions with t.Helper()
func assertContains(t *testing.T, output, expected string) {
	t.Helper()
	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain %q, got:\n%s", expected, output)
	}
}

// ❌ INCORRECT - Vague error messages
if !strings.Contains(output, expected) {
	t.Error("test failed")  // Not helpful!
}
```

---

## Build and Test Workflows

### Running Tests

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./cmd

# Run specific test
go test ./cmd -run TestBRTag

# Verbose output
go test -v ./...

# With coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Building

```bash
# Build binary
go build -o mermaid-ascii .

# Build with version info
go build -ldflags "-X main.version=1.0.0" -o mermaid-ascii .

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o mermaid-ascii-linux-amd64 .
GOOS=darwin GOARCH=amd64 go build -o mermaid-ascii-darwin-amd64 .
GOOS=windows GOARCH=amd64 go build -o mermaid-ascii-windows-amd64.exe .
```

---

## Common Gotchas

### 1. Pointer vs Value Receivers

```go
// ✅ CORRECT - Use pointer receiver for methods that modify state
func (g *graph) setLabelLines() {
	g.labelWidth = calculateWidth()  // Modifies graph
}

// ✅ CORRECT - Use value receiver for read-only methods
func (c Config) IsValid() bool {
	return c.MaxWidth >= 0  // Doesn't modify config
}
```

### 2. String Immutability

**Remember:** Strings in Go are immutable. Use `strings.Builder` for efficient concatenation.

### 3. Slice Append

```go
// ✅ CORRECT - Assign result back
lines = append(lines, newLine)

// ❌ INCORRECT - Missing assignment
append(lines, newLine)  // Does nothing!
```

### 4. Range Loop Variables

```go
// ✅ CORRECT - Use index or create copy
for i := range tests {
	t.Run(tests[i].name, func(t *testing.T) {
		// Use tests[i]
	})
}

// ⚠️ WARNING - Loop variable capture (Go < 1.22)
for _, tt := range tests {
	t.Run(tt.name, func(t *testing.T) {
		// tt may be reused across iterations in older Go versions
		// Fixed in Go 1.22+
	})
}
```

### 5. Error Wrapping

```go
// ✅ CORRECT - Use %w to wrap errors (enables errors.Is/As)
return fmt.Errorf("failed to parse: %w", err)

// ❌ INCORRECT - Use %v (loses error chain)
return fmt.Errorf("failed to parse: %v", err)
```

---

## Code Review Practices

### Review Modified Files

**Always review all modified files for errors, omissions, anti-patterns, or other issues before finalizing changes:**

- **Errors**: Syntax errors, logic bugs, unhandled errors, type mismatches
- **Omissions**: Missing tests, incomplete implementations, missing error handling
- **Anti-patterns**: 
  - Missing error checks
  - Inefficient string operations
  - Non-table-driven tests
  - Missing test cases
  - Unvalidated configuration
- **Design Issues**: Poor naming, missing documentation, unclear logic
- **Performance Issues**: Inefficient algorithms, unnecessary allocations, repeated work

**Use systematic review process:**
1. Check each modified file for completeness
2. Verify all errors are handled
3. Ensure tests exist and pass
4. Validate documentation is up-to-date
5. Look for edge cases and boundary conditions
6. Confirm Go idioms are followed

---

## Documentation Standards

### Function Documentation

All exported functions must have godoc comments:

```go
// RenderDiagram converts Mermaid diagram syntax to ASCII art.
// It returns the rendered ASCII output and any error encountered.
//
// The input should be valid Mermaid syntax. The config parameter
// controls rendering options like width, padding, and style.
//
// Example:
//
//	diagram := "graph LR\nA --> B"
//	output, err := RenderDiagram(diagram, config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(output)
func RenderDiagram(input string, config *Config) (string, error) {
	// Implementation
}
```

### Package Documentation

```go
// Package cmd implements the command-line interface and core
// rendering logic for converting Mermaid diagrams to ASCII art.
//
// This package provides diagram parsing, layout calculation,
// and ASCII rendering with support for various Mermaid diagram types.
package cmd
```

---

## Git Commit Messages

### Summarizing Changes

**When preparing a commit message, briefly summarize all changed files using a small number of high-level bullet points:**

```bash
# ✅ CORRECT - High-level summary
feat: Add HTML tag support for multi-line labels

- Add <br/> and <br> tag conversion in graph rendering
- Update text wrapping to split on converted newlines
- Add comprehensive test coverage with table-driven tests
- Update documentation with usage examples

# ❌ INCORRECT - Too detailed or file-by-file
Update cmd/graph.go
Update cmd/graph_test.go
Update internal/diagram/config.go
...
```

**Guidelines:**
- **Use high-level themes** instead of listing individual file changes
- **Group related changes** into conceptual bullet points (3-5 bullets)
- **Focus on user-facing changes** and their benefits
- **Include context** about why changes were made when relevant
- Review modified files to ensure all changes are represented

---

## Agent Skills

This project includes Agent Skills in `.github/skills/` for common procedural patterns.

### Available Skills

1. **#go-testing** - Create and maintain Go unit tests with table-driven patterns
   - **Use when**: Adding tests for new/modified functions
   - Covers table-driven tests, helper functions, test organization

2. **#go-build-and-test** - Build, test, and validate Go code
   - **Use when**: Building binaries, running test suites, checking coverage
   - Covers build flags, cross-compilation, test execution

3. **#code-review** - Systematically review modified files before finalizing changes
   - **Use when**: Before committing, after completing edits, or preparing pull requests
   - Checks for errors, omissions, anti-patterns, design issues

4. **#git-commit-message** - Generate concise, thematic commit messages
   - **Use when**: Preparing commits with multiple file changes
   - Creates high-level summaries grouped by theme

### When to Use Agent Skills

**Invoke skills explicitly** (using `#skill-name` in your message) when:
- You need step-by-step guidance through a multi-step procedural task
- The pattern is well-defined and documented in a skill file
- You want the agent to follow a specific structured approach
- You're less familiar with a particular pattern

**Skills are automatically selected** when:
- Your request clearly matches a skill's purpose
- Copilot recognizes the task fits a documented skill pattern
- No explicit skill reference is needed for straightforward requests

---

## Cross-Project Development

**When making changes from other project directories, always check for and use project-specific guidance:**

- **`.github/copilot-instructions.md`** - Project-specific instructions and anti-patterns
- **`.github/skills/`** - Agent Skills with procedural patterns

These files contain critical project-specific context including:
- Language-specific patterns (R, Python, Go, etc.)
- Testing standards and code review requirements
- Common gotcas and error patterns
- Development workflows

---

## Project Structure

```
mermaid-ascii/
├── cmd/                    # CLI commands and core rendering
│   ├── graph.go           # Graph diagram rendering
│   ├── graph_test.go      # Graph tests
│   ├── parse.go           # Mermaid syntax parsing
│   ├── root.go            # Cobra CLI root command
│   └── ...
├── internal/
│   ├── diagram/           # Configuration and rendering infrastructure
│   │   ├── config.go
│   │   └── config_test.go
│   └── sequence/          # Sequence diagram rendering
├── main.go                # Entry point
├── go.mod                 # Go module definition
├── go.sum                 # Dependency checksums
└── README.md              # Documentation
```

---

## Development Workflow

1. **Make changes** - Modify code with clear intent
2. **Write tests** - Table-driven tests alongside code
3. **Run tests** - `go test ./...`
4. **Review code** - Use #code-review skill
5. **Build** - `go build -o mermaid-ascii .`
6. **Manual test** - Test CLI with sample diagrams
7. **Commit** - Use #git-commit-message for message

---

## References

- **Go Documentation**: https://go.dev/doc/
- **Go Testing**: https://go.dev/doc/tutorial/add-a-test
- **Cobra CLI**: https://github.com/spf13/cobra
- **Mermaid Syntax**: https://mermaid.js.org/intro/
- **Table-Driven Tests**: https://go.dev/wiki/TableDrivenTests
