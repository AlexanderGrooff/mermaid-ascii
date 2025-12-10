package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlexanderGrooff/mermaid-ascii/internal/diagram/testutil"
	log "github.com/sirupsen/logrus"
)

func verifyMap(t *testing.T, testCaseFile string) {
	tc, err := testutil.ReadTestCase(testCaseFile)
	if err != nil {
		t.Fatalf("Failed to read test case file: %v", err)
	}

	properties, err := mermaidFileToMap(tc.Mermaid, "cli")
	if err != nil {
		log.Fatal("Failed to parse mermaid: ", err)
	}
	properties.paddingX = tc.PaddingX
	properties.paddingY = tc.PaddingY
	actualMap := drawMap(properties)
	if tc.Expected != actualMap {
		expectedWithSpaces := testutil.VisualizeWhitespace(tc.Expected)
		actualWithSpaces := testutil.VisualizeWhitespace(actualMap)
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

// Sequence diagram tests moved to sequence_test.go
