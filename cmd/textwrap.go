package cmd

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

func wrapLabelLines(text string, width int) []string {
	lines := splitLabelLines(text)
	if width <= 0 {
		return lines
	}
	wrapped := []string{}
	for _, line := range lines {
		wrapped = append(wrapped, wrapLine(line, width)...)
	}
	if len(wrapped) == 0 {
		return []string{""}
	}
	return wrapped
}

func splitLabelLines(text string) []string {
	if text == "" {
		return []string{""}
	}
	return strings.Split(text, "\n")
}

func wrapLine(line string, width int) []string {
	if width <= 0 || runewidth.StringWidth(line) <= width {
		return []string{line}
	}
	words := strings.Fields(line)
	if len(words) == 0 {
		return []string{""}
	}
	lines := []string{}
	current := ""
	currentWidth := 0
	for _, word := range words {
		wordWidth := runewidth.StringWidth(word)
		if current == "" {
			if wordWidth <= width {
				current = word
				currentWidth = wordWidth
				continue
			}
			parts := hardWrapWord(word, width)
			if len(parts) > 1 {
				lines = append(lines, parts[:len(parts)-1]...)
			}
			current = parts[len(parts)-1]
			currentWidth = runewidth.StringWidth(current)
			continue
		}
		if currentWidth+1+wordWidth <= width {
			current += " " + word
			currentWidth += 1 + wordWidth
			continue
		}
		lines = append(lines, current)
		current = ""
		currentWidth = 0
		if wordWidth <= width {
			current = word
			currentWidth = wordWidth
			continue
		}
		parts := hardWrapWord(word, width)
		if len(parts) > 1 {
			lines = append(lines, parts[:len(parts)-1]...)
		}
		current = parts[len(parts)-1]
		currentWidth = runewidth.StringWidth(current)
	}
	if current != "" {
		lines = append(lines, current)
	}
	if len(lines) == 0 {
		return []string{""}
	}
	return lines
}

func hardWrapWord(word string, width int) []string {
	if width <= 0 || runewidth.StringWidth(word) <= width {
		return []string{word}
	}
	parts := []string{}
	runes := []rune(word)
	currentPart := []rune{}
	currentWidth := 0
	
	for _, r := range runes {
		runeW := runewidth.RuneWidth(r)
		if currentWidth+runeW > width && len(currentPart) > 0 {
			parts = append(parts, string(currentPart))
			currentPart = []rune{r}
			currentWidth = runeW
		} else {
			currentPart = append(currentPart, r)
			currentWidth += runeW
		}
	}
	if len(currentPart) > 0 {
		parts = append(parts, string(currentPart))
	}
	if len(parts) == 0 {
		return []string{""}
	}
	return parts
}

func maxLineWidth(lines []string) int {
	maxWidth := 0
	for _, line := range lines {
		lineWidth := runewidth.StringWidth(line)
		if lineWidth > maxWidth {
			maxWidth = lineWidth
		}
	}
	return maxWidth
}
