# Mermaid ASCII

Render mermaid diagrams in your terminal:

## Installation

You can download the binary from Github releases:

```bash
# Get the latest release
$ curl -s https://api.github.com/repos/AlexanderGrooff/mermaid-ascii/releases/latest | grep "browser_download_url.*mermaid-ascii" | grep "$(uname)_$(uname -m)" | cut -d: -f2,3 | tr -d \" | wget -qi -
# Unzip it
$ tar xvzf mermaid-ascii_*.tar.gz
$ ./mermaid-ascii --help
```

You can also build it yourself:

```bash
$ git clone
$ cd mermaid-ascii
$ go build
$ mermaid-ascii --help
```

## Usage
```bash
$ cat test.mermaid
A --> C
A --> D
B --> C
C --> D
B --> D
$ mermaid-ascii --file test.mermaid
+---+    +---+    +---+
|   |    |   |    |   |
| A |--->| C |<---| B |
|   |    |   |    |   |
+---+    +---+    +---+
  |        |        |
  |        |        |
  |        v        |
  |      +---+      |
  |      |   |      |
  +----->| D |<-----+
         |   |
         +---+
$ mermaid-ascii --file test.mermaid -x 8 -y 8
+---+       +---+       +---+
|   |       |   |       |   |
| A |------>| C |<------| B |
|   |       |   |       |   |
+---+       +---+       +---+
  |           |           |
  |           |           |
  |           |           |
  |           |           |
  |           |           |
  |           |           |
  |           v           |
  |         +---+         |
  |         |   |         |
  +-------->| D |<--------+
            |   |
            +---+
$ mermaid-ascii -f ./test.mermaid -p 3
+-------+    +-------+    +-------+
|       |    |       |    |       |
|       |    |       |    |       |
|       |    |       |    |       |
|   A   |--->|   C   |<---|   B   |
|       |    |       |    |       |
|       |    |       |    |       |
|       |    |       |    |       |
+-------+    +-------+    +-------+
    |            |            |
    |            |            |
    |            v            |
    |        +-------+        |
    |        |       |        |
    |        |       |        |
    |        |       |        |
    +------->|   D   |<-------+
             |       |
             |       |
             |       |
             +-------+
$ mermaid-ascii --help
Generate ASCII diagrams from mermaid code.

Usage:
  mermaid-ascii [flags]

Flags:
  -f, --file string    Mermaid file to parse
  -h, --help           help for mermaid-ascii
  -x, --paddingX int   Horizontal space between nodes (default 5)
  -y, --paddingY int   Vertical space between nodes (default 4)
  -v, --verbose        verbose output
```

## TODOs

The baseline components for Mermaid work, but there are a lot of things that are not supported yet. Here's a list of things that are not yet supported:

### Syntax support

- [ ] Graph directions like `graph LR` and `graph TB`
- [ ] `classDef` and `class`
- [ ] `subgraph`
- [ ] Shapes other than rectangles
- [ ] `A & B`
- [ ] Multiple arrows on one line
- [ ] Whitespacing and comments

### Rendering

- [ ] Prevent arrows overlapping nodes
- [ ] Support for multiline nodes
- [ ] Place nodes in a more compact way
- [ ] Prevent rendering more than X characters wide (like 80 for terminal width)
