package sequence

// BoxChars defines the characters used for drawing the diagram.
type BoxChars struct {
	TopLeft      rune
	TopRight     rune
	BottomLeft   rune
	BottomRight  rune
	Horizontal   rune
	Vertical     rune
	TeeDown      rune
	TeeRight     rune
	TeeLeft      rune
	Cross        rune
	ArrowRight   rune
	ArrowLeft    rune
	CrossHead    rune // head of -x / --x (lost/failed message)
	PointRight   rune // head of -) / --) (async message)
	PointLeft    rune
	Circle       rune // central-connection marker ("()") on a lifeline
	SolidLine    rune
	DottedLine   rune
	SelfTopRight rune
	SelfBottom   rune
	// Active* draw a lifeline during an activation period: a heavy stroke
	// standing in for mermaid's activation box, with matching junctions where
	// messages attach to or cross it.
	ActiveVertical rune
	ActiveTeeRight rune
	ActiveTeeLeft  rune
	ActiveCross    rune
}

var ASCII = BoxChars{
	TopLeft:      '+',
	TopRight:     '+',
	BottomLeft:   '+',
	BottomRight:  '+',
	Horizontal:   '-',
	Vertical:     '|',
	TeeDown:      '+',
	TeeRight:     '+',
	TeeLeft:      '+',
	Cross:        '+',
	ArrowRight:   '>',
	ArrowLeft:    '<',
	CrossHead:    'x',
	PointRight:   ')',
	PointLeft:    '(',
	Circle:       'o',
	SolidLine:    '-',
	DottedLine:   '.',
	SelfTopRight: '+',
	SelfBottom:   '+',
	// ASCII has no heavy stroke; '#' is the conventional filled marker.
	ActiveVertical: '#',
	ActiveTeeRight: '#',
	ActiveTeeLeft:  '#',
	ActiveCross:    '#',
}

var Unicode = BoxChars{
	TopLeft:     '┌',
	TopRight:    '┐',
	BottomLeft:  '└',
	BottomRight: '┘',
	Horizontal:  '─',
	Vertical:    '│',
	TeeDown:     '┬',
	TeeRight:    '├',
	TeeLeft:     '┤',
	Cross:       '┼',
	ArrowRight:  '►',
	ArrowLeft:   '◄',
	CrossHead:   '×',
	PointRight:  ')',
	PointLeft:   '(',
	// 'o' rather than '○': the latter is East-Asian-ambiguous width and
	// would break column alignment in CJK-capable terminals.
	Circle:       'o',
	SolidLine:    '─',
	DottedLine:   '┈',
	SelfTopRight: '┐',
	SelfBottom:   '┘',
	// Heavy vertical, with heavy-vertical/light-horizontal junctions so an
	// arrow attaching to an active lifeline still reads as a light arrow.
	ActiveVertical: '┃',
	ActiveTeeRight: '┠',
	ActiveTeeLeft:  '┨',
	ActiveCross:    '╂',
}
