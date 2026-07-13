// Package er parses and renders mermaid entity-relationship (erDiagram)
// diagrams as ASCII: entity attribute tables connected by crow's-foot
// relationships.
package er

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/AlexanderGrooff/mermaid-ascii/pkg/diagram"
)

const erKeyword = "erDiagram"

// Cardinality is one end of a relationship (crow's-foot notation).
type Cardinality int

const (
	OnlyOne    Cardinality = iota // ||   exactly one
	ZeroOrOne                     // |o / o|   zero or one
	ZeroOrMore                    // }o / o{   zero or more
	OneOrMore                     // }| / |{   one or more
)

// Attribute is one row of an entity's attribute table.
type Attribute struct {
	Type    string
	Name    string
	Keys    []string // PK, FK, UK
	Comment string
}

// Entity is a named box with an optional list of attributes.
type Entity struct {
	Name       string
	Attributes []Attribute
}

// Relationship connects two entities with a cardinality at each end.
type Relationship struct {
	Left, Right string
	LeftCard    Cardinality
	RightCard   Cardinality
	Identifying bool // true for a solid (--) line, false for dashed (..)
	Label       string
}

// ErDiagram is a parsed entity-relationship diagram. Entities are kept in first
// -seen order; a relationship referencing an undeclared entity auto-creates it.
type ErDiagram struct {
	Entities      []*Entity
	Relationships []*Relationship
	byName        map[string]*Entity
}

var (
	// leftCard/rightCard map the crow's-foot tokens to a Cardinality.
	leftCard  = map[string]Cardinality{"||": OnlyOne, "|o": ZeroOrOne, "}o": ZeroOrMore, "}|": OneOrMore}
	rightCard = map[string]Cardinality{"||": OnlyOne, "o|": ZeroOrOne, "o{": ZeroOrMore, "|{": OneOrMore}

	// relationshipRegex matches "ENTITY1 <lcard><line><rcard> ENTITY2 : label"
	// where line is -- (identifying), .. / -. / .- (non-identifying).
	relationshipRegex = regexp.MustCompile(
		`^\s*(?:"([^"]+)"|(\S+))\s+([|}o]{2})(--|\.\.|-\.|\.-)([|o{]{2})\s+(?:"([^"]+)"|(\S+))\s*:\s*(.*)$`)

	// textRelRegex matches the word-alias form: "E1 <cardphrase> to|optionally to
	// <cardphrase> E2 : label" (e.g. "PERSON many optionally to one CAR : owns").
	textRelRegex = regexp.MustCompile(
		`^\s*(\S+)\s+(.+?)\s+(optionally to|to)\s+(.+?)\s+(\S+)\s*:\s*(.*)$`)

	// cardPhrase maps mermaid's word/short cardinality aliases to a Cardinality.
	cardPhrase = map[string]Cardinality{
		"zero or one": ZeroOrOne, "one or zero": ZeroOrOne,
		"zero or more": ZeroOrMore, "zero or many": ZeroOrMore, "0+": ZeroOrMore,
		"many": ZeroOrMore, "many(0)": ZeroOrMore,
		"one or more": OneOrMore, "one or many": OneOrMore, "1+": OneOrMore, "many(1)": OneOrMore,
		"only one": OnlyOne, "one": OnlyOne, "1": OnlyOne,
	}

	// styleLineRegex matches visual-styling / accessibility lines that carry no
	// ASCII meaning; they're skipped so a stray one doesn't fail a diagram.
	styleLineRegex = regexp.MustCompile(`(?i)^\s*(classDef|class|style|accTitle|accDescr)\b`)

	// entityHeaderRegex matches the opening of an attribute block: `NAME {`.
	entityHeaderRegex = regexp.MustCompile(`^\s*(?:"([^"]+)"|(\S+))\s*\{\s*$`)

	// loneEntityRegex matches an entity declared on its own with no block/relation.
	loneEntityRegex = regexp.MustCompile(`^\s*(?:"([^"]+)"|([A-Za-z0-9_-]+))\s*$`)

	// attrKeyRegex matches a PK/FK/UK key token (possibly comma-separated).
	attrKeyRegex = regexp.MustCompile(`^(?:PK|FK|UK)(?:\s*,\s*(?:PK|FK|UK))*$`)
)

// IsErDiagram reports whether the input's first meaningful line declares an
// erDiagram (case-insensitive, whole token).
func IsErDiagram(input string) bool {
	for _, line := range strings.Split(input, "\n") {
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "%%") {
			continue
		}
		low := strings.ToLower(t)
		return low == strings.ToLower(erKeyword) ||
			strings.HasPrefix(low, strings.ToLower(erKeyword)+" ")
	}
	return false
}

func (d *ErDiagram) entity(name string) *Entity {
	if e, ok := d.byName[name]; ok {
		return e
	}
	e := &Entity{Name: name}
	d.byName[name] = e
	d.Entities = append(d.Entities, e)
	return e
}

// Parse parses an erDiagram into entities and relationships.
func Parse(input string) (*ErDiagram, error) {
	lines := diagram.RemoveComments(diagram.SplitLines(strings.TrimSpace(input)))
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty input")
	}
	if !IsErDiagram(strings.TrimSpace(lines[0])) {
		return nil, fmt.Errorf("expected %q keyword", erKeyword)
	}

	d := &ErDiagram{byName: map[string]*Entity{}}

	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		// Visual styling / accessibility lines carry no ASCII meaning — skip.
		if styleLineRegex.MatchString(line) {
			continue
		}

		// Entity attribute block: NAME { ... }
		if m := entityHeaderRegex.FindStringSubmatch(line); m != nil {
			name := firstNonEmpty(m[1], m[2])
			e := d.entity(name)
			attrs, next, err := parseAttributeBlock(lines, i+1)
			if err != nil {
				return nil, fmt.Errorf("entity %q: %w", name, err)
			}
			e.Attributes = append(e.Attributes, attrs...) // multiple blocks accumulate
			i = next                                      // index of the closing "}"
			continue
		}

		// Relationship: ENTITY1 ||--o{ ENTITY2 : label
		if m := relationshipRegex.FindStringSubmatch(line); m != nil {
			left := firstNonEmpty(m[1], m[2])
			right := firstNonEmpty(m[6], m[7])
			d.entity(left)
			d.entity(right)
			d.Relationships = append(d.Relationships, &Relationship{
				Left:        left,
				Right:       right,
				LeftCard:    leftCard[m[3]],
				RightCard:   rightCard[m[5]],
				Identifying: m[4] == "--",
				Label:       strings.TrimSpace(m[8]),
			})
			continue
		}

		// Relationship, word-alias form: E1 <cardphrase> to <cardphrase> E2 : lbl
		if m := textRelRegex.FindStringSubmatch(line); m != nil {
			lc, lok := cardPhrase[strings.ToLower(m[2])]
			rc, rok := cardPhrase[strings.ToLower(m[4])]
			if lok && rok {
				d.entity(m[1])
				d.entity(m[5])
				d.Relationships = append(d.Relationships, &Relationship{
					Left:        m[1],
					Right:       m[5],
					LeftCard:    lc,
					RightCard:   rc,
					Identifying: strings.EqualFold(m[3], "to"),
					Label:       strings.TrimSpace(m[6]),
				})
				continue
			}
		}

		// A bare entity name declares an entity with no attributes.
		if m := loneEntityRegex.FindStringSubmatch(line); m != nil {
			d.entity(firstNonEmpty(m[1], m[2]))
			continue
		}

		return nil, fmt.Errorf("line %d: invalid syntax: %q", i+1, line)
	}

	if len(d.Entities) == 0 {
		return nil, fmt.Errorf("no entities found")
	}
	return d, nil
}

// parseAttributeBlock reads attribute rows until the closing "}", returning the
// attributes and the index of the closing-brace line.
func parseAttributeBlock(lines []string, start int) ([]Attribute, int, error) {
	var attrs []Attribute
	for i := start; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if line == "}" {
			return attrs, i, nil
		}
		attr, err := parseAttribute(line)
		if err != nil {
			return nil, i, err
		}
		attrs = append(attrs, attr)
	}
	return nil, len(lines), fmt.Errorf("unclosed attribute block (missing '}')")
}

// parseAttribute parses "type name [keys] [\"comment\"]".
func parseAttribute(line string) (Attribute, error) {
	// Pull a trailing quoted comment off first.
	comment := ""
	if idx := strings.Index(line, `"`); idx != -1 {
		if end := strings.LastIndex(line, `"`); end > idx {
			comment = line[idx+1 : end]
			line = strings.TrimSpace(line[:idx])
		}
	}
	fields := splitAttrTokens(line)
	if len(fields) < 2 {
		return Attribute{}, fmt.Errorf("attribute needs a type and name: %q", line)
	}
	attr := Attribute{Type: fields[0], Name: fields[1], Comment: comment}
	if rest := strings.TrimSpace(strings.Join(fields[2:], " ")); rest != "" {
		if !attrKeyRegex.MatchString(rest) {
			return Attribute{}, fmt.Errorf("unexpected attribute tokens %q", rest)
		}
		for _, k := range strings.Split(rest, ",") {
			attr.Keys = append(attr.Keys, strings.TrimSpace(k))
		}
	}
	return attr, nil
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// splitAttrTokens splits on whitespace but keeps parenthesised groups intact, so
// a type like "decimal(10, 2)" or "varchar(255)" stays a single token.
func splitAttrTokens(s string) []string {
	var toks []string
	var cur strings.Builder
	depth := 0
	flush := func() {
		if cur.Len() > 0 {
			toks = append(toks, cur.String())
			cur.Reset()
		}
	}
	for _, r := range s {
		switch {
		case r == '(':
			depth++
			cur.WriteRune(r)
		case r == ')':
			if depth > 0 {
				depth--
			}
			cur.WriteRune(r)
		case (r == ' ' || r == '\t') && depth == 0:
			flush()
		default:
			cur.WriteRune(r)
		}
	}
	flush()
	return toks
}
