package cmd

import (
	"strings"
	"unicode/utf8"
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
	if width <= 0 || utf8.RuneCountInString(line) <= width {
		return []string{line}
	}
	words := strings.Fields(line)
	if len(words) == 0 {
		return []string{""}
	}
	lines := []string{}
	current := ""
	currentLen := 0
	for _, word := range words {
		wordLen := utf8.RuneCountInString(word)
		if current == "" {
			if wordLen <= width {
				current = word
				currentLen = wordLen
				continue
			}
			parts := hardWrapWord(word, width)
			if len(parts) > 1 {
				lines = append(lines, parts[:len(parts)-1]...)
			}
			current = parts[len(parts)-1]
			currentLen = utf8.RuneCountInString(current)
			continue
		}
		if currentLen+1+wordLen <= width {
			current += " " + word
			currentLen += 1 + wordLen
			continue
		}
		lines = append(lines, current)
		current = ""
		currentLen = 0
		if wordLen <= width {
			current = word
			currentLen = wordLen
			continue
		}
		parts := hardWrapWord(word, width)
		if len(parts) > 1 {
			lines = append(lines, parts[:len(parts)-1]...)
		}
		current = parts[len(parts)-1]
		currentLen = utf8.RuneCountInString(current)
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
	if width <= 0 || utf8.RuneCountInString(word) <= width {
		return []string{word}
	}
	parts := []string{}
	runes := []rune(word)
	for len(runes) > width {
		parts = append(parts, string(runes[:width]))
		runes = runes[width:]
	}
	word = string(runes)
	if word != "" {
		parts = append(parts, word)
	}
	if len(parts) == 0 {
		return []string{""}
	}
	return parts
}

func maxLineWidth(lines []string) int {
	maxWidth := 0
	for _, line := range lines {
		lineWidth := utf8.RuneCountInString(line)
		if lineWidth > maxWidth {
			maxWidth = lineWidth
		}
	}
	return maxWidth
}
