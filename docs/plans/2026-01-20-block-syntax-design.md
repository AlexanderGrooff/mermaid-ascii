# Sequence Diagram Block Syntax Support

## Overview

Add support for Mermaid's block syntax in sequence diagrams. Blocks group related messages and can be nested arbitrarily.

## Supported Block Types

| Block | Syntax | Purpose |
|-------|--------|---------|
| `loop` | `loop [label]` ... `end` | Repeated actions |
| `alt` | `alt [label]` ... `else [label]` ... `end` | Conditional branches |
| `opt` | `opt [label]` ... `end` | Optional block |
| `par` | `par [label]` ... `and [label]` ... `end` | Parallel execution |
| `critical` | `critical [label]` ... `option [label]` ... `end` | Critical section with options |
| `break` | `break [label]` ... `end` | Break out of loop |
| `rect` | `rect [color]` ... `end` | Highlight region (color ignored) |

All keywords are case-insensitive.

## Visual Examples

### Loop
```
     │         │
     │  ┌──────┴───────────────────┐
     │  │ loop Video Stream        │
     │  │  ┌───────────────────────┤
     │  │  │ H.264 Frame           │
     │  ├──┼──────────────────────►│
     │  │  │                       │
     │  │  │ H.264 Frame           │
     │◄─┼──┼───────────────────────┤
     │  └──┴───────────────────────┘
     │         │
```

### Alt/Else
```
     │  ┌──────────────────────┐
     │  │ alt Success          │
     │  │  ┌───────────────────┤
     │  │  │ Response: 200     │
     │◄─┼──┼───────────────────┤
     │  ├──┼───────────────────┤
     │  │  │ else Error        │
     │  │  ├───────────────────┤
     │  │  │ Response: 500     │
     │◄─┼──┼───────────────────┤
     │  └──┴───────────────────┘
```

### Par (Parallel)
```
     │         │         │
     │  ┌──────┴─────────┴──────┐
     │  │ par Control Commands  │
     │  │  ┌────────────────────┤
     │  │  │ Touch Event        │
     │  ├──┼───────────────────►│
     │  ├──┼────────────────────┤
     │  │  │ and Events         │
     │  │  ├────────────────────┤
     │  │  │ Clipboard Changed  │
     │  │  │◄───────────────────┤
     │  └──┴────────────────────┘
```

## Data Model

### New Types

```go
type BlockType int

const (
    BlockLoop BlockType = iota
    BlockAlt
    BlockOpt
    BlockPar
    BlockCritical
    BlockBreak
    BlockRect
)

type Block struct {
    Type     BlockType
    Label    string
    Sections []*BlockSection
}

type BlockSection struct {
    Label    string
    Elements []DiagramElement
}

func (*Block) isElement() {}
```

### Block Type Characteristics

| Block | Divider Keywords | Min Sections | Max Sections |
|-------|------------------|--------------|--------------|
| `loop` | - | 1 | 1 |
| `alt` | `else` | 1 | N |
| `opt` | - | 1 | 1 |
| `par` | `and` | 1 | N |
| `critical` | `option` | 1 | N |
| `break` | - | 1 | 1 |
| `rect` | - | 1 | 1 |

## Parser Design

### Regexes

```go
blockStartRegex   = regexp.MustCompile(`(?i)^\s*(loop|alt|opt|par|critical|break|rect)\s*(.*)$`)
blockDividerRegex = regexp.MustCompile(`(?i)^\s*(else|and|option)\s*(.*)$`)
blockEndRegex     = regexp.MustCompile(`(?i)^\s*end\s*$`)
```

### Recursive Descent Parser

```go
func (sd *SequenceDiagram) parseBlock(lines []string, startIdx int, participants map[string]*Participant) (*Block, int, error) {
    // 1. Parse block start line to get type and label
    // 2. Create block with first section
    // 3. Loop through lines:
    //    - If nested block start: recurse, add to current section
    //    - If divider: validate for block type, start new section
    //    - If end: return completed block
    //    - Otherwise: parse as message/note, add to current section
    // 4. Return block and index after 'end'
}
```

### Divider Validation

| Block Type | Valid Dividers |
|------------|----------------|
| `alt` | `else` |
| `par` | `and` |
| `critical` | `option` |
| others | none (error if divider found) |

## Renderer Design

### Entry Point

```go
func renderBlock(block *Block, layout *diagramLayout, chars BoxChars, depth int) []string {
    // 1. Find participant range (leftmost/rightmost used in block)
    // 2. Calculate box bounds with indent based on depth
    // 3. Render top border with block type and label
    // 4. For each section:
    //    a. Render section elements recursively
    //    b. If not last, render divider line with section label
    // 5. Render bottom border
}
```

### Helper Functions

```go
func findBlockParticipantRange(block *Block) (minIdx, maxIdx int)
func renderBlockBorder(label string, isTop bool, ...) string
func renderBlockDivider(label string, ...) string
```

### Nesting

The `depth` parameter tracks nesting level:
- `depth=0`: outermost block, minimal indent
- `depth=1`: one level nested, additional indent
- etc.

Each nesting level adds ~2 characters of left indent to avoid overlapping borders.

## Testing Strategy

1. **Parser tests**: Each block type, multi-section, nesting, case insensitivity, invalid dividers
2. **Renderer tests**: Box drawing, dividers, nested blocks, mixed content
3. **Integration**: `scratch/note-syntax.mermaid` with loop and par

## Backward Compatibility

- Existing diagrams without blocks work unchanged
- Blocks integrate with existing messages and notes via `DiagramElement` interface
- Autonumber applies only to messages, not block structure

## Future Considerations

- `activate`/`deactivate` for participant activation bars
- `box` for grouping participants
- Background colors for `rect` blocks (would require terminal color support)
