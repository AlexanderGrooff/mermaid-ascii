package graph

import (
	"regexp"
	"strings"

	"github.com/mattn/go-runewidth"
)

var htmlBreakPattern = regexp.MustCompile(`(?i)<br\s*/?>`)

const graphLabelLineGap = 1

type graphLabel struct {
	lines []string
	width int
}

func newGraphLabel(raw string) graphLabel {
	normalized := htmlBreakPattern.ReplaceAllString(raw, "\n")
	normalized = strings.ReplaceAll(normalized, `\n`, "\n")

	lines := strings.Split(normalized, "\n")
	if len(lines) == 0 {
		lines = []string{""}
	}

	width := 0
	for _, line := range lines {
		width = Max(width, runewidth.StringWidth(line))
	}

	return graphLabel{
		lines: lines,
		width: width,
	}
}

func (l graphLabel) height() int {
	return len(l.lines)
}

func (l graphLabel) contentHeight() int {
	if len(l.lines) == 0 {
		return 0
	}
	return len(l.lines) + (len(l.lines)-1)*graphLabelLineGap
}
