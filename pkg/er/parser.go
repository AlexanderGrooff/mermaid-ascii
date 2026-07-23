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

	// styleLineRegex matches visual-styling lines that carry no ASCII meaning;
	// they're skipped so a stray one doesn't fail a diagram. Checked after the
	// entity-header form so an entity that happens to be named `class` still works.
	styleLineRegex = regexp.MustCompile(`(?i)^\s*(classDef|class|style)\b`)

	// accLineRegex matches accessibility metadata: `accTitle: …`, `accDescr: …`,
	// or the multi-line `accDescr {` block form (whose body is skipped too).
	accLineRegex = regexp.MustCompile(`(?i)^\s*(accTitle|accDescr)\s*[:{]`)

	// entityHeaderRegex matches the opening of an attribute block, with an
	// optional alias: `NAME {`, `NAME alias {`, `NAME[Alias] {`, or
	// `NAME["Alias Label"] {`.
	entityHeaderRegex = regexp.MustCompile(`^\s*(?:"([^"]+)"|([^\s{}["]+))(?:\s*\[\s*"?([^"\]]+?)"?\s*\]|\s+(\S+))?\s*\{\s*$`)

	// loneEntityRegex matches an entity declared on its own (no block/relation),
	// with an optional alias: `NAME`, `NAME alias`, `NAME[Alias]`, or
	// `NAME["Alias Label"]`.
	loneEntityRegex = regexp.MustCompile(`^\s*(?:"([^"]+)"|([^\s{}:|"\[]+))(?:\s*\[\s*"?([^"\]]+?)"?\s*\]|\s+(\S+))?\s*$`)

	// attrKeyRegex matches a PK/FK/UK key token (possibly comma-separated).
	attrKeyRegex = regexp.MustCompile(`^(?:PK|FK|UK)(?:\s*,\s*(?:PK|FK|UK))*$`)

	// emptyBlockRegex matches a one-line empty attribute block suffix: `NAME {}`.
	emptyBlockRegex = regexp.MustCompile(`\s*\{\s*\}\s*$`)

	// classShorthandRegex matches a `:::class[,class…]` styling decoration on an
	// entity; like classDef/class lines it carries no ASCII meaning.
	classShorthandRegex = regexp.MustCompile(`:::[\w,-]+`)

	// subgraphRegex matches the opening of an er subgraph block (an unreleased
	// upstream feature this renderer rejects rather than mis-draws).
	subgraphRegex = regexp.MustCompile(`^subgraph\b`)
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
	if !IsErDiagram(input) {
		return nil, fmt.Errorf("expected %q keyword", erKeyword)
	}
	// Comments are stripped in place (not filtered out as whole lines) so error
	// messages report the caller's real line numbers.
	lines := diagram.SplitLines(strings.TrimSpace(input))
	for i, l := range lines {
		lines[i] = stripComment(l)
	}

	d := &ErDiagram{byName: map[string]*Entity{}}

	seenKeyword := false
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if !seenKeyword { // the erDiagram keyword line itself (verified above)
			seenKeyword = true
			continue
		}

		// Accessibility / layout-direction lines carry no ASCII meaning here —
		// skip them (including an `accDescr { … }` block) rather than failing.
		if accLineRegex.MatchString(line) {
			if strings.HasSuffix(line, "{") {
				for i++; i < len(lines) && !strings.Contains(lines[i], "}"); i++ {
				}
			}
			continue
		}
		if directionRegex.MatchString(line) {
			continue
		}

		// er subgraphs (unreleased upstream) would otherwise misparse into
		// bogus entity boxes — reject them loudly instead.
		if subgraphRegex.MatchString(line) || line == "end" {
			return nil, fmt.Errorf("line %d: er subgraphs are not supported", i+1)
		}

		// `:::class` styling decorations carry no ASCII meaning; strip them
		// (outside quoted strings) so the decorated statement parses normally.
		if strings.Contains(line, ":::") {
			line = stripClassShorthand(line)
		}

		// A one-line empty attribute block (`NAME {}`) is just an entity
		// declaration; strip the block and let the lone-entity form match.
		if emptyBlockRegex.MatchString(line) {
			line = strings.TrimSpace(emptyBlockRegex.ReplaceAllString(line, ""))
		}

		// Entity attribute block: NAME { ... } (with optional alias). Checked
		// before the style-line skip so entities named e.g. `class` still work.
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
		// Checked before the style-line skip — a relationship always carries a
		// connector and a colon, which no styling directive does, so an entity
		// named `class` keeps its relationships.
		if d.parseRelationship(line) {
			continue
		}

		// Visual styling directives (classDef/class/style) have no ASCII meaning.
		if styleLineRegex.MatchString(line) {
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

	// A statement-less erDiagram is valid mermaid; it renders as empty output.
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
		// Mermaid allows the closing brace on the last attribute's line
		// ("string title}"), so a trailing "}" closes the block after the
		// attribute (if any) on that line is parsed. Quoted comments never
		// end in a bare "}" — the quote is the last character.
		closes := strings.HasSuffix(line, "}")
		if closes {
			line = strings.TrimSpace(strings.TrimSuffix(line, "}"))
		}
		if line != "" {
			attr, err := parseAttribute(line)
			if err != nil {
				return nil, i, fmt.Errorf("line %d: %w", i+1, err)
			}
			attrs = append(attrs, attr)
		}
		if closes {
			return attrs, i, nil
		}
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
	// Backtick escaping (`geo.accuracy`) exists only to smuggle special
	// characters past mermaid's lexer; the backticks themselves never render.
	attr := Attribute{
		Type:    strings.Trim(fields[0], "`"),
		Name:    strings.Trim(fields[1], "`"),
		Comment: comment,
	}
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

// stripClassShorthand removes `:::class[,class…]` decorations from the parts
// of a line outside double-quoted strings.
func stripClassShorthand(line string) string {
	parts := strings.Split(line, `"`)
	for i := 0; i < len(parts); i += 2 { // even indices are outside quotes
		parts[i] = classShorthandRegex.ReplaceAllString(parts[i], "")
	}
	return strings.Join(parts, `"`)
}

// stripComment drops a %% comment (whole-line or trailing) from a line.
// %% inside a quoted string (a label or attribute comment) is kept, matching
// mermaid's lexer, which tokenizes strings before comments.
func stripComment(line string) string {
	inQuote := false
	for i := 0; i < len(line); i++ {
		switch {
		case line[i] == '"':
			inQuote = !inQuote
		case !inQuote && line[i] == '%' && i+1 < len(line) && line[i+1] == '%':
			return strings.TrimRight(line[:i], " \t")
		}
	}
	return line
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
	// Collapse label whitespace like mermaid's SVG text rendering does; a
	// whitespace-only label would otherwise punch blank holes in its line.
	label := strings.Join(strings.Fields(strings.Trim(strings.TrimSpace(line[colon+1:]), `"`)), " ")

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
// the entity id and the cardinality text. Quoted names may contain spaces.
func splitEntityCard(part string, entityFirst bool) (entity, card string) {
	part = strings.TrimSpace(part)
	if entityFirst {
		if strings.HasPrefix(part, `"`) {
			if end := strings.Index(part[1:], `"`); end >= 0 {
				return part[1 : end+1], strings.TrimSpace(part[end+2:])
			}
		}
		toks := strings.Fields(part)
		if len(toks) == 0 {
			return "", ""
		}
		return strings.Trim(toks[0], `"`), strings.Join(toks[1:], " ")
	}
	if strings.HasSuffix(part, `"`) {
		if start := strings.LastIndex(part[:len(part)-1], `"`); start >= 0 {
			return part[start+1 : len(part)-1], strings.TrimSpace(part[:start])
		}
	}
	toks := strings.Fields(part)
	if len(toks) == 0 {
		return "", ""
	}
	return strings.Trim(toks[len(toks)-1], `"`), strings.Join(toks[:len(toks)-1], " ")
}

// splitAttrTokens splits on whitespace but keeps parenthesised and
// backtick-escaped groups intact, so a type like "decimal(10, 2)" or a name
// like `two words` stays a single token.
func splitAttrTokens(s string) []string {
	var toks []string
	var cur strings.Builder
	depth := 0
	inTick := false
	flush := func() {
		if cur.Len() > 0 {
			toks = append(toks, cur.String())
			cur.Reset()
		}
	}
	for _, r := range s {
		switch {
		case r == '`':
			inTick = !inTick
			cur.WriteRune(r)
		case r == '(' && !inTick:
			depth++
			cur.WriteRune(r)
		case r == ')' && !inTick:
			if depth > 0 {
				depth--
			}
			cur.WriteRune(r)
		case (r == ' ' || r == '\t') && depth == 0 && !inTick:
			flush()
		default:
			cur.WriteRune(r)
		}
	}
	flush()
	return toks
}
