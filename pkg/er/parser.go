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

// Entity is a named box with an optional list of attributes. Name is the id
// used in relationships; Display is the label shown in the box (an alias if one
// was given, otherwise the name).
type Entity struct {
	Name       string
	Display    string
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
	// cardAny maps every cardinality form mermaid accepts — crow's-foot tokens
	// (either side), numeric shorthands, and word phrases — to a Cardinality.
	cardAny = map[string]Cardinality{
		// crow's-foot tokens (accepted on either side)
		"||": OnlyOne,
		"|o": ZeroOrOne, "o|": ZeroOrOne,
		"}o": ZeroOrMore, "o{": ZeroOrMore,
		"}|": OneOrMore, "|{": OneOrMore,
		// numeric / word shorthands
		"1": OnlyOne, "only one": OnlyOne, "one": OnlyOne,
		"zero or one": ZeroOrOne, "one or zero": ZeroOrOne,
		"0+": ZeroOrMore, "zero or more": ZeroOrMore, "zero or many": ZeroOrMore,
		"many": ZeroOrMore, "many(0)": ZeroOrMore,
		"1+": OneOrMore, "one or more": OneOrMore, "one or many": OneOrMore, "many(1)": OneOrMore,
	}

	// lineOpRegex matches the relationship line operator (a 2-char run of - and
	// . — entity names use single dashes, so this uniquely marks the connector).
	lineOpRegex = regexp.MustCompile(`[-.]{2}`)

	// directionRegex matches the optional "direction TB|LR|..." layout directive.
	directionRegex = regexp.MustCompile(`(?i)^\s*direction\s+\S+\s*$`)

	// styleLineRegex matches visual-styling / accessibility lines that carry no
	// ASCII meaning; they're skipped so a stray one doesn't fail a diagram.
	styleLineRegex = regexp.MustCompile(`(?i)^\s*(classDef|class|style|accTitle|accDescr)\b`)

	// entityHeaderRegex matches the opening of an attribute block, with an
	// optional alias: `NAME {`, `NAME alias {`, or `NAME["Alias Label"] {`.
	entityHeaderRegex = regexp.MustCompile(`^\s*(?:"([^"]+)"|(\S+?))(?:\s*\["([^"]+)"\]|\s+(\S+))?\s*\{\s*$`)

	// loneEntityRegex matches an entity declared on its own (no block/relation),
	// with an optional alias: `NAME`, `NAME alias`, or `NAME["Alias Label"]`.
	loneEntityRegex = regexp.MustCompile(`^\s*(?:"([^"]+)"|([A-Za-z0-9_.-]+))(?:\s*\["([^"]+)"\]|\s+(\S+))?\s*$`)

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
	e := &Entity{Name: name, Display: name}
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

		// Visual styling / accessibility / layout-direction lines carry no ASCII
		// meaning here — skip them rather than failing the diagram.
		if styleLineRegex.MatchString(line) || directionRegex.MatchString(line) {
			continue
		}

		// Entity attribute block: NAME { ... } (with optional alias)
		if m := entityHeaderRegex.FindStringSubmatch(line); m != nil {
			name := firstNonEmpty(m[1], m[2])
			e := d.entity(name)
			if alias := firstNonEmpty(m[3], m[4]); alias != "" {
				e.Display = alias
			}
			attrs, next, err := parseAttributeBlock(lines, i+1)
			if err != nil {
				return nil, fmt.Errorf("entity %q: %w", name, err)
			}
			e.Attributes = append(e.Attributes, attrs...) // multiple blocks accumulate
			i = next                                      // index of the closing "}"
			continue
		}

		// Relationship (any cardinality form: crow's-foot, numeric, or words).
		if d.parseRelationship(line) {
			continue
		}

		// A bare entity name (with optional alias) declares an entity.
		if m := loneEntityRegex.FindStringSubmatch(line); m != nil {
			e := d.entity(firstNonEmpty(m[1], m[2]))
			if alias := firstNonEmpty(m[3], m[4]); alias != "" {
				e.Display = alias
			}
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
		if line == "" || strings.HasPrefix(line, "//") {
			continue // blank or a // note line (some diagrams use these)
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
	// Pull a trailing quoted comment off first. Tolerate an unclosed quote
	// (some hand-written diagrams forget the closing ") by taking the rest.
	comment := ""
	if idx := strings.Index(line, `"`); idx != -1 {
		if end := strings.LastIndex(line, `"`); end > idx {
			comment = line[idx+1 : end]
		} else {
			comment = line[idx+1:]
		}
		line = strings.TrimSpace(line[:idx])
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

// parseRelationship parses any relationship form and appends it to d, returning
// true if the line was a relationship. Cardinality on each side may be a
// crow's-foot token (||, o{, …), a numeric shorthand (1, 0+, 1+), or a word
// phrase (only one, one or more, …); the connector is -- / .. / -. / .- or the
// word operator "to" / "optionally to".
func (d *ErDiagram) parseRelationship(line string) bool {
	colon := strings.Index(line, ":")
	if colon < 0 {
		return false
	}
	main := strings.TrimSpace(line[:colon])
	label := strings.Trim(strings.TrimSpace(line[colon+1:]), `"`)

	var left, right string
	var identifying bool
	if loc := lineOpRegex.FindStringIndex(main); loc != nil {
		left = strings.TrimSpace(main[:loc[0]])
		right = strings.TrimSpace(main[loc[1]:])
		identifying = main[loc[0]:loc[1]] == "--"
	} else if i, w := findWordOp(main); i >= 0 {
		left = strings.TrimSpace(main[:i])
		right = strings.TrimSpace(main[i+len(w):])
		identifying = w == " to "
	} else {
		return false
	}

	e1, lcard := splitEntityCard(left, true)
	e2, rcard := splitEntityCard(right, false)
	lc, lok := cardAny[strings.ToLower(lcard)]
	rc, rok := cardAny[strings.ToLower(rcard)]
	if e1 == "" || e2 == "" || !lok || !rok {
		return false
	}
	d.entity(e1)
	d.entity(e2)
	d.Relationships = append(d.Relationships, &Relationship{
		Left: e1, Right: e2, LeftCard: lc, RightCard: rc,
		Identifying: identifying, Label: label,
	})
	return true
}

// findWordOp locates the " to " / " optionally to " word connector.
func findWordOp(s string) (int, string) {
	for _, w := range []string{" optionally to ", " to "} {
		if i := strings.Index(s, w); i >= 0 {
			return i, w
		}
	}
	return -1, ""
}

// splitEntityCard splits "ENTITY <card>" (entityFirst) or "<card> ENTITY" into
// the entity id and the cardinality text.
func splitEntityCard(part string, entityFirst bool) (entity, card string) {
	toks := strings.Fields(part)
	if len(toks) == 0 {
		return "", ""
	}
	if entityFirst {
		return strings.Trim(toks[0], `"`), strings.Join(toks[1:], " ")
	}
	return strings.Trim(toks[len(toks)-1], `"`), strings.Join(toks[:len(toks)-1], " ")
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
