package sequence

// BoxChars defines the characters used for drawing the diagram.
type BoxChars struct {
	TopLeft      rune
	TopRight     rune
	BottomLeft   rune
	BottomRight  rune
	Horizontal   rune
	Vertical     rune
	TeeUp        rune
	TeeDown      rune
	TeeRight     rune
	TeeLeft      rune
	Cross        rune
	ArrowRight   rune
	ArrowLeft    rune
	SolidLine    rune
	DottedLine   rune
	SelfTopRight rune
	SelfBottom   rune

	RoundedTopLeft     rune
	RoundedTopRight    rune
	RoundedBottomLeft  rune
	RoundedBottomRight rune

	DottedHorizontal rune
	DottedVertical   rune

	DoubleHorizontal   rune
	DoubleVertical     rune
	DoubleTopLeft      rune
	DoubleTopRight     rune
	DoubleBottomLeft   rune
	DoubleBottomRight  rune
	DoubleTeeRight     rune
	DoubleTeeLeft      rune
	DoubleTeeUp        rune
	DoubleTeeDown      rune
	DoubleCross        rune

	MixedTopLeft     rune
	MixedTopRight    rune
	MixedBottomLeft  rune
	MixedBottomRight rune
	MixedTeeRight    rune
	MixedTeeLeft     rune
}

var ASCII = BoxChars{
	TopLeft:      '+',
	TopRight:     '+',
	BottomLeft:   '+',
	BottomRight:  '+',
	Horizontal:   '-',
	Vertical:     '|',
	TeeUp:        '+',
	TeeDown:      '+',
	TeeRight:     '+',
	TeeLeft:      '+',
	Cross:        '+',
	ArrowRight:   '>',
	ArrowLeft:    '<',
	SolidLine:    '-',
	DottedLine:   '.',
	SelfTopRight: '+',
	SelfBottom:   '+',

	RoundedTopLeft:     '+',
	RoundedTopRight:    '+',
	RoundedBottomLeft:  '+',
	RoundedBottomRight: '+',

	DottedHorizontal: '.',
	DottedVertical:   ':',

	DoubleHorizontal:  '=',
	DoubleVertical:    '#',
	DoubleTopLeft:     '#',
	DoubleTopRight:    '#',
	DoubleBottomLeft:  '#',
	DoubleBottomRight: '#',
	DoubleTeeRight:    '#',
	DoubleTeeLeft:     '#',
	DoubleTeeUp:       '#',
	DoubleTeeDown:     '#',
	DoubleCross:       '#',

	MixedTopLeft:     '+',
	MixedTopRight:    '+',
	MixedBottomLeft:  '+',
	MixedBottomRight: '+',
	MixedTeeRight:    '+',
	MixedTeeLeft:     '+',
}

var Unicode = BoxChars{
	TopLeft:      '┌',
	TopRight:     '┐',
	BottomLeft:   '└',
	BottomRight:  '┘',
	Horizontal:   '─',
	Vertical:     '│',
	TeeUp:        '┴',
	TeeDown:      '┬',
	TeeRight:     '├',
	TeeLeft:      '┤',
	Cross:        '┼',
	ArrowRight:   '►',
	ArrowLeft:    '◄',
	SolidLine:    '─',
	DottedLine:   '┈',
	SelfTopRight: '┐',
	SelfBottom:   '┘',

	RoundedTopLeft:     '╭',
	RoundedTopRight:    '╮',
	RoundedBottomLeft:  '╰',
	RoundedBottomRight: '╯',

	DottedHorizontal: '┄',
	DottedVertical:   '┆',

	DoubleHorizontal:  '═',
	DoubleVertical:    '║',
	DoubleTopLeft:     '╔',
	DoubleTopRight:    '╗',
	DoubleBottomLeft:  '╚',
	DoubleBottomRight: '╝',
	DoubleTeeRight:    '╠',
	DoubleTeeLeft:     '╣',
	DoubleTeeUp:       '╩',
	DoubleTeeDown:     '╦',
	DoubleCross:       '╬',

	MixedTopLeft:     '╓',
	MixedTopRight:    '╖',
	MixedBottomLeft:  '╙',
	MixedBottomRight: '╜',
	MixedTeeRight:    '╟',
	MixedTeeLeft:     '╢',
}

type BlockBoxChars struct {
	TopLeft     rune
	TopRight    rune
	BottomLeft  rune
	BottomRight rune
	Horizontal  rune
	Vertical    rune
	TeeRight    rune
	TeeLeft     rune
	TeeUp       rune
	TeeDown     rune
	Cross       rune
}

func GetBlockChars(blockType BlockType, base BoxChars) BlockBoxChars {
	switch blockType {
	case BlockLoop:
		return BlockBoxChars{
			TopLeft:     base.RoundedTopLeft,
			TopRight:    base.RoundedTopRight,
			BottomLeft:  base.RoundedBottomLeft,
			BottomRight: base.RoundedBottomRight,
			Horizontal:  base.Horizontal,
			Vertical:    base.Vertical,
			TeeRight:    base.TeeRight,
			TeeLeft:     base.TeeLeft,
			TeeUp:       base.TeeUp,
			TeeDown:     base.TeeDown,
			Cross:       base.Cross,
		}
	case BlockAlt, BlockOpt:
		return BlockBoxChars{
			TopLeft:     base.TopLeft,
			TopRight:    base.TopRight,
			BottomLeft:  base.BottomLeft,
			BottomRight: base.BottomRight,
			Horizontal:  base.DottedHorizontal,
			Vertical:    base.DottedVertical,
			TeeRight:    base.TeeRight,
			TeeLeft:     base.TeeLeft,
			TeeUp:       base.TeeUp,
			TeeDown:     base.TeeDown,
			Cross:       base.Cross,
		}
	case BlockPar:
		return BlockBoxChars{
			TopLeft:     base.MixedTopLeft,
			TopRight:    base.MixedTopRight,
			BottomLeft:  base.MixedBottomLeft,
			BottomRight: base.MixedBottomRight,
			Horizontal:  base.Horizontal,
			Vertical:    base.DoubleVertical,
			TeeRight:    base.MixedTeeRight,
			TeeLeft:     base.MixedTeeLeft,
			TeeUp:       base.TeeUp,
			TeeDown:     base.TeeDown,
			Cross:       base.Cross,
		}
	case BlockCritical, BlockBreak:
		return BlockBoxChars{
			TopLeft:     base.DoubleTopLeft,
			TopRight:    base.DoubleTopRight,
			BottomLeft:  base.DoubleBottomLeft,
			BottomRight: base.DoubleBottomRight,
			Horizontal:  base.DoubleHorizontal,
			Vertical:    base.DoubleVertical,
			TeeRight:    base.DoubleTeeRight,
			TeeLeft:     base.DoubleTeeLeft,
			TeeUp:       base.DoubleTeeUp,
			TeeDown:     base.DoubleTeeDown,
			Cross:       base.DoubleCross,
		}
	default:
		return BlockBoxChars{
			TopLeft:     base.TopLeft,
			TopRight:    base.TopRight,
			BottomLeft:  base.BottomLeft,
			BottomRight: base.BottomRight,
			Horizontal:  base.Horizontal,
			Vertical:    base.Vertical,
			TeeRight:    base.TeeRight,
			TeeLeft:     base.TeeLeft,
			TeeUp:       base.TeeUp,
			TeeDown:     base.TeeDown,
			Cross:       base.Cross,
		}
	}
}
