package cmd

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
)

type testCase struct {
	mermaid  string
	expected string
	paddingX int
	paddingY int
}

func readTestCase(filePath string) (testCase, error) {
	tc := testCase{paddingX: 5, paddingY: 5}

	file, err := os.Open(filePath)
	if err != nil {
		return tc, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var mermaid, expectedMap strings.Builder
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
						return tc, convErr
					}
					if strings.EqualFold(match[1], "paddingX") {
						tc.paddingX = paddingValue
					} else {
						tc.paddingY = paddingValue
					}
					continue
				}
			}
			mermaidStarted = true
			mermaid.WriteString(line + "\n")
		} else {
			expectedMap.WriteString(line + "\n")
		}
	}

	tc.mermaid = mermaid.String()
	tc.expected = strings.TrimSuffix(expectedMap.String(), "\n")
	return tc, scanner.Err()
}

func verifyMap(t *testing.T, testCaseFile string) {
	tc, err := readTestCase(testCaseFile)
	if err != nil {
		t.Fatalf("Failed to read test case file: %v", err)
	}

	properties, err := mermaidFileToMap(tc.mermaid, "cli")
	if err != nil {
		log.Fatal("Failed to parse mermaid: ", err)
	}
	properties.paddingX = tc.paddingX
	properties.paddingY = tc.paddingY
	actualMap := drawMap(properties)
	if tc.expected != actualMap {
		expectedWithSpaces := strings.ReplaceAll(tc.expected, " ", "·")
		actualWithSpaces := strings.ReplaceAll(actualMap, " ", "·")
		t.Errorf("Map didn't match actual map\nExpected:\n%v\nActual:\n%v", expectedWithSpaces, actualWithSpaces)
	}
}

func TestASCII(t *testing.T) {
	useAscii = true
	dir := "testdata/ascii"
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("Failed to read directory %s: %v", dir, err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".txt") {
			t.Run(file.Name(), func(t *testing.T) {
				verifyMap(t, filepath.Join(dir, file.Name()))
			})
		}
	}
}

func TestExtendedChars(t *testing.T) {
	useAscii = false
	dir := "testdata/extended-chars"
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("Failed to read directory %s: %v", dir, err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".txt") {
			t.Run(file.Name(), func(t *testing.T) {
				verifyMap(t, filepath.Join(dir, file.Name()))
			})
		}
	}
}
