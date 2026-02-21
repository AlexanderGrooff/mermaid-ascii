package testutil

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// TestCase represents a test case for diagram rendering.
// It contains the input Mermaid syntax and expected output.
type TestCase struct {
	Mermaid  string
	Expected string
	PaddingX int
	PaddingY int
}

// ReadTestCase reads a test case file with optional padding configuration.
// File format:
//
//	[paddingX = N]  // optional
//	[paddingY = N]  // optional
//	<mermaid code>
//	---
//	<expected output>
//
// For sequence diagrams, use ReadSequenceTestCase which doesn't support padding config.
func ReadTestCase(filePath string) (*TestCase, error) {
	tc := &TestCase{PaddingX: 5, PaddingY: 5}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var mermaid, expected strings.Builder
	inMermaid := true
	mermaidStarted := false
	paddingRegex := regexp.MustCompile(`^(?i)(padding[xy])\s*=\s*(\d+)\s*$`)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			inMermaid = false
			continue
		}
		if inMermaid {
			trimmed := strings.TrimSpace(line)
			if !mermaidStarted {
				if trimmed == "" {
					continue
				}
				if match := paddingRegex.FindStringSubmatch(trimmed); match != nil {
					paddingValue, convErr := strconv.Atoi(match[2])
					if convErr != nil {
						return nil, convErr
					}
					if strings.EqualFold(match[1], "paddingX") {
						tc.PaddingX = paddingValue
					} else {
						tc.PaddingY = paddingValue
					}
					continue
				}
			}
			mermaidStarted = true
			mermaid.WriteString(line + "\n")
		} else {
			expected.WriteString(line + "\n")
		}
	}

	tc.Mermaid = mermaid.String()
	tc.Expected = strings.TrimSuffix(expected.String(), "\n")
	return tc, scanner.Err()
}

// ReadSequenceTestCase reads a test case file for sequence diagrams.
// Uses more precise "\n---\n" separator to avoid matching "---" in ASCII box borders.
// File format:
//
//	<mermaid code>
//	---
//	<expected output>
func ReadSequenceTestCase(filePath string) (*TestCase, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Split on "\n---\n" to avoid matching "---" in ASCII box borders
	parts := strings.Split(string(content), "\n---\n")
	if len(parts) != 2 {
		return nil, fmt.Errorf("test case file must have exactly one '---' separator (on its own line)")
	}

	return &TestCase{
		Mermaid:  strings.TrimSpace(parts[0]),
		Expected: strings.TrimSpace(parts[1]),
		PaddingX: 5,
		PaddingY: 5,
	}, nil
}

// NormalizeWhitespace removes trailing spaces and empty lines for comparison.
// This is useful for comparing expected vs actual output where trailing whitespace doesn't matter.
func NormalizeWhitespace(s string) string {
	lines := strings.Split(s, "\n")
	var normalized []string
	for _, line := range lines {
		trimmed := strings.TrimRight(line, " ")
		if trimmed != "" || len(normalized) > 0 {
			normalized = append(normalized, trimmed)
		}
	}
	// Remove trailing empty lines
	for len(normalized) > 0 && normalized[len(normalized)-1] == "" {
		normalized = normalized[:len(normalized)-1]
	}
	return strings.Join(normalized, "\n")
}

// VisualizeWhitespace replaces spaces with · for debugging test failures.
func VisualizeWhitespace(s string) string {
	return strings.ReplaceAll(s, " ", "·")
}
