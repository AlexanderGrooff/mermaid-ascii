package cmd

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
)

func readTestCase(filePath string) (string, string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var mermaid, expectedMap strings.Builder
	inMermaid := true

	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			inMermaid = false
			continue
		}
		if inMermaid {
			mermaid.WriteString(line + "\n")
		} else {
			expectedMap.WriteString(line + "\n")
		}
	}

	return mermaid.String(), strings.TrimSuffix(expectedMap.String(), "\n"), scanner.Err()
}

func verifyMap(t *testing.T, testCaseFile string) {
	mermaid, expectedMap, err := readTestCase(testCaseFile)
	if err != nil {
		t.Fatalf("Failed to read test case file: %v", err)
	}

	properties, err := mermaidFileToMap(mermaid, "cli")
	if err != nil {
		log.Fatal("Failed to parse mermaid: ", err)
	}
	actualMap := drawMap(properties)
	if expectedMap != actualMap {
		expectedWithSpaces := strings.ReplaceAll(expectedMap, " ", "·")
		actualWithSpaces := strings.ReplaceAll(actualMap, " ", "·")
		t.Errorf("Map didn't match actual map\nExpected:\n%v\nActual:\n%v", expectedWithSpaces, actualWithSpaces)
	}
}

func TestASCII(t *testing.T) {
	useExtendedChars = false
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
	useExtendedChars = true
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
