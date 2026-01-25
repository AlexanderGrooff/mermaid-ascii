# Sequence Diagram Note Syntax Support

## Overview

Add support for Mermaid's note syntax in sequence diagrams. Notes are annotations that appear alongside the message flow to provide context or explanations.

## Supported Syntax

Four variants will be supported:

```mermaid
Note over Actor: text           # Note above a single actor
Note over Actor1,Actor2: text   # Note spanning multiple actors  
Note left of Actor: text        # Note positioned left of actor
Note right of Actor: text       # Note positioned right of actor
```

All variants are case-insensitive (`Note`, `NOTE`, `note`).

## Visual Examples

Given participants `Client`, `ESA`, `WS`, `Phone`:

### Note over single actor
```
    │             │          │            │
┌───┴───────────────┐
│ Extract control   │
│ server            │
└───┬───────────────┘
    │             │          │            │
```

### Note spanning multiple actors
```
    │             │          │            │
┌───┴─────────────┴──────────┴──┐
│ Spanning note                 │
└───┬─────────────┬──────────┬──┘
    │             │          │            │
```

### Note left of actor
```
    │             │          │            │
┌─────────┐
│ Left    ├──────┤
│ note    │      │
└─────────┘      │
    │             │          │            │
```

### Note right of actor
```
    │             │          │            │
                                   ┌──────────┐
                             │─────┤ Right    │
                             │     │ note     │
                             │     └──────────┘
    │             │          │            │
```

## Data Model

### New Types

```go
type NotePosition int

const (
    NoteOver    NotePosition = iota  // Note over Actor or Note over Actor1,Actor2
    NoteLeftOf                        // Note left of Actor
    NoteRightOf                       // Note right of Actor
)

type Note struct {
    Position    NotePosition
    Actors      []*Participant  // 1 actor for left/right, 1-2 for "over"
    Text        string
}
```

### Modified Types

```go
type DiagramElement interface{}  // Messages and Notes both implement this

type SequenceDiagram struct {
    Participants []*Participant
    Messages     []*Message       // Keep for backward compatibility
    Elements     []DiagramElement // Ordered sequence of messages + notes
    Autonumber   bool
}
```

## Parser Design

Add regex to match all note variants:

```go
noteRegex = regexp.MustCompile(`(?i)^\s*note\s+(over|left\s+of|right\s+of)\s+([^:]+):\s*(.*)$`)
```

The `parseNote` method will:
1. Match the regex
2. Determine position from capture group 1
3. Parse actor(s) from capture group 2 (comma-separated for multi-actor)
4. Extract text from capture group 3
5. Look up or auto-create participants
6. Append Note to Elements slice

## Renderer Design

### Entry Point

```go
func renderNote(note *Note, layout *diagramLayout, chars BoxChars) []string {
    switch note.Position {
    case NoteOver:
        return renderNoteOver(note, layout, chars)
    case NoteLeftOf:
        return renderNoteLeftOf(note, layout, chars)
    case NoteRightOf:
        return renderNoteRightOf(note, layout, chars)
    }
    return nil
}
```

### Rendering Logic

- **NoteOver**: Calculate horizontal bounds from actor center(s), draw box with text, lifelines connect at box edges using `┴`/`┬` connectors
- **NoteLeftOf**: Draw box left of actor's lifeline, connect with horizontal line
- **NoteRightOf**: Draw box right of actor's lifeline, connect with horizontal line

### Main Loop Change

The `Render` function will iterate over `Elements` instead of just `Messages`, dispatching to `renderNote` or `renderMessage` as appropriate.

## Testing Strategy

1. **Parser tests**: Each note variant, case insensitivity, auto-participant creation
2. **Renderer tests**: Golden-file tests for each note position type
3. **Integration test**: Use `scratch/note-syntax.mermaid` as real-world validation

## Backward Compatibility

- Keep `Messages` field populated for any code depending on it
- Existing diagrams without notes continue to work unchanged
- Autonumber only applies to messages, not notes

## Future Considerations

- Multi-line note text (currently single line only)
- Note styling/theming
- Note text wrapping at configurable width
