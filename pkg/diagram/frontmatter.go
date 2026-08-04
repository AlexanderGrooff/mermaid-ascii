package diagram

import "strings"

// StripFrontmatter removes a leading YAML frontmatter block (delimited by
// `---` lines) from a mermaid document, returning the remaining input and the
// frontmatter's title, if one was set.
//
// Mermaid uses frontmatter for a diagram title and theme/config overrides
// (colors, CSS). The config has no meaning in ASCII output, so it is
// discarded; the title is surfaced so callers can print it above the diagram,
// as mermaid does.
//
// Matching mermaid's own frontmatter semantics (frontmatter.spec.ts):
// frontmatter is only recognised at the start of the document, the closing
// delimiter must sit at the same indentation as the opening one, and an
// unclosed block is not frontmatter at all — the input is returned untouched
// for the diagram parser to deal with.
func StripFrontmatter(input string) (rest string, title string) {
	lines := strings.Split(input, "\n")

	start := 0
	for start < len(lines) && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	if start >= len(lines) || !isDelimiter(lines[start]) {
		return input, ""
	}
	indent := lines[start][:strings.Index(lines[start], "-")]

	for i := start + 1; i < len(lines); i++ {
		if isDelimiter(lines[i]) && lines[i][:strings.Index(lines[i], "-")] == indent {
			return strings.Join(lines[i+1:], "\n"), title
		}
		// Only a top-level `title:` key counts — indented occurrences are
		// nested config values (e.g. inside themeCSS), not the title. YAML
		// requires whitespace after the colon for a mapping ("title:xyz" is a
		// plain scalar, not a key), and an unquoted value ends at a comment.
		trimmed := strings.TrimRight(lines[i], " \t\r")
		if v, ok := strings.CutPrefix(trimmed, indent+"title:"); ok && (v == "" || v[0] == ' ' || v[0] == '\t') {
			v = strings.TrimSpace(v)
			if !strings.HasPrefix(v, `"`) && !strings.HasPrefix(v, `'`) {
				if idx := strings.Index(v, " #"); idx != -1 {
					v = strings.TrimSpace(v[:idx])
				}
			}
			title = strings.Trim(v, `"'`)
		}
	}
	return input, ""
}

// isDelimiter reports whether a line is a frontmatter delimiter: `---` with
// optional surrounding whitespace.
func isDelimiter(line string) bool {
	return strings.TrimSpace(line) == "---"
}
