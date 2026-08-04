package diagram

import "testing"

// Cases mirror mermaid's own frontmatter.spec.ts semantics.
func TestStripFrontmatter(t *testing.T) {
	cases := []struct {
		name      string
		in        string
		wantRest  string
		wantTitle string
	}{
		{"no frontmatter", "graph TD\nA-->B", "graph TD\nA-->B", ""},
		{"title extracted", "---\ntitle: foo\n---\ndiagram", "diagram", "foo"},
		{"empty frontmatter", "---\n\n---\ndiagram", "diagram", ""},
		{"unclosed is not frontmatter", "---\ntitle: foo\ndiagram", "---\ntitle: foo\ndiagram", ""},
		{"only at document start", "diagram\n---\ntitle: foo\n---", "diagram\n---\ntitle: foo\n---", ""},
		{"delimiter inside a value line", "---\ntitle: foo---bar\n---\ndiagram", "diagram", "foo---bar"},
		{"unknown keys ignored", "---\nconfig:\n  theme: dark\ninvalid: x\n---\ndiagram", "diagram", ""},
		{"quoted title", "---\ntitle: \"Customers service\"\n---\ndiagram", "diagram", "Customers service"},
		{"boolean-looking title stays text", "---\ntitle: true\n---\ndiagram", "diagram", "true"},
		{"leading blank lines", "\n\n---\ntitle: t\n---\ndiagram", "diagram", "t"},
		{"matching indented delimiters", "  ---\n  title: t\n  ---\ndiagram", "diagram", "t"},
		{"mismatched closing indent not a close", "---\ntitle: t\n   ---\ndiagram", "---\ntitle: t\n   ---\ndiagram", ""},
		{"indented title is not the title", "---\nconfig:\n  title: nested\n---\ndiagram", "diagram", ""},
		{"crlf input", "---\r\ntitle: t\r\n---\r\ndiagram", "diagram", "t"},
		{"multiline themeCSS config", "---\nconfig:\n  themeCSS: |\n    rect { fill: red; }\n---\nerDiagram", "erDiagram", ""},
		{"inline comment stripped from title", "---\ntitle: hi # note\n---\ndiagram", "diagram", "hi"},
		{"hash kept inside quoted title", "---\ntitle: \"a # b\"\n---\ndiagram", "diagram", "a # b"},
		{"no space after colon is not a mapping", "---\ntitle:xyz\n---\ndiagram", "diagram", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rest, title := StripFrontmatter(c.in)
			if rest != c.wantRest {
				t.Errorf("rest = %q, want %q", rest, c.wantRest)
			}
			if title != c.wantTitle {
				t.Errorf("title = %q, want %q", title, c.wantTitle)
			}
		})
	}
}
