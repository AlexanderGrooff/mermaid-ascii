package cmd

import (
	"bufio"
	"os"
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

func TestSingleNode(t *testing.T) {
	verifyMap(t, "testdata/single_node.txt")
}

func TestSingleNodeLongerName(t *testing.T) {
	verifyMap(t, "testdata/single_node_longer_name.txt")
}

func TestTwoSingleRootNodes(t *testing.T) {
	verifyMap(t, "testdata/two_single_root_nodes.txt")
}

func TestTwoNodesLinked(t *testing.T) {
	verifyMap(t, "testdata/two_nodes_linked.txt")
}

func TestTwoNodesLongerNames(t *testing.T) {
	verifyMap(t, "testdata/two_nodes_longer_names.txt")
}

func TestTwoLayerSingleGraph(t *testing.T) {
	verifyMap(t, "testdata/two_layer_single_graph.txt")
}

func TestBacklinkFromTop(t *testing.T) {
	verifyMap(t, "testdata/backlink_from_top.txt")
}

func TestBacklinkFromBottom(t *testing.T) {
	verifyMap(t, "testdata/backlink_from_bottom.txt")
}

func TestTwoLayerSingleGraphLongerNames(t *testing.T) {
	verifyMap(t, "testdata/two_layer_single_graph_longer_names.txt")
}

func TestThreeNodes(t *testing.T) {
	verifyMap(t, "testdata/three_nodes.txt")
}

func TestThreeNodesSingleLine(t *testing.T) {
	verifyMap(t, "testdata/three_nodes_single_line.txt")
}

func TestTwoRootNodes(t *testing.T) {
	verifyMap(t, "testdata/two_root_nodes.txt")
}

func TestTwoRootNodesLongerNames(t *testing.T) {
	verifyMap(t, "testdata/two_root_nodes_longer_names.txt")
}

func TestAmpersandLHS(t *testing.T) {
	verifyMap(t, "testdata/ampersand_lhs.txt")
}

func TestAmpersandRHS(t *testing.T) {
	verifyMap(t, "testdata/ampersand_rhs.txt")
}

func TestAmpersandLHSAndRHS(t *testing.T) {
	verifyMap(t, "testdata/ampersand_lhs_and_rhs.txt")
}

func TestAmpersandWithoutEdge(t *testing.T) {
	verifyMap(t, "testdata/ampersand_without_edge.txt")
}

func TestSelfReference(t *testing.T) {
	verifyMap(t, "testdata/self_reference.txt")
}

func TestSelfReferenceWithEdge(t *testing.T) {
	verifyMap(t, "testdata/self_reference_with_edge.txt")
}

func TestBackReferenceFromChild(t *testing.T) {
	verifyMap(t, "testdata/back_reference_from_child.txt")
}

func TestPreserveOrderOfDefinition(t *testing.T) {
	verifyMap(t, "testdata/preserve_order_of_definition.txt")
}
