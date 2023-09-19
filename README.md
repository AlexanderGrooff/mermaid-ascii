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
A --> B
A --> C
B --> C
B --> D
C --> D
$ mermaid-ascii --file test.mermaid
+---+          +---+          +---+
|   |          |   |          |   |
| A |--------->| B |--------->| D |
|   |          |   |          |   |
+---+          +---+          +---+
  \              |              ^
   \             |             /
    \            v            /
     \         +---+         /
      \        |   |        /
       ------->| C | ------/
               |   |
               +---+

# Increase horizontal spacing
$ mermaid-ascii --file test.mermaid -x 8
+---+                +---+                +---+
|   |                |   |                |   |
| A |--------------->| B |--------------->| D |
|   |                |   |                |   |
+---+                +---+                +---+
  \                    |                    ^
   \                   |                   /
    \                  v                  /
     \               +---+               /
      \              |   |              /
       ------------->| C | ------------/
                     |   |
                     +---+

# Increase box padding
$ mermaid-ascii -f ./test.mermaid -p 3
+-----+          +-----+          +-----+
|     |          |     |          |     |
|     |          |     |          |     |
|  A  |--------->|  B  |--------->|  D  |
|     |          |     |          |     |
|     |          |     |          |     |
+-----+          +-----+          +-----+
   \                v                ^
    \            +-----+            /
     \           |     |           /
      \          |     |          /
       --------->|  C  | --------/
                 |     |
                 |     |
                 +-----+

# Labeled edges
$ cat test.mermaid
A --> B
A --> C
B --> C
B -->|example| D
C --> D
$ mermaid-ascii -f ./test.mermaid -x 2 -y 4
+---+    +---+           +---+
|   |    |   |           |   |
| A |--->| B |--example->| D |
|   |    |   |           |   |
+---+    +---+           +---+
  \        |               ^
   \       |              /
    \      v             /
     \   +---+          /
      \  |   |         /
       ->| C | -------/
         |   |
         +---+
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

- [x] Labelled edges (like `A -->|label| B`)
- [ ] Graph directions like `graph LR` and `graph TB`
- [ ] `classDef` and `class`
- [ ] `subgraph`
- [ ] Shapes other than rectangles
- [ ] `A & B`
- [ ] Multiple arrows on one line (like `A --> B --> C`)
- [ ] Whitespacing and comments

### Rendering

- [x] Diagonal arrows
- [ ] Prevent arrows overlapping nodes
- [ ] Place nodes in a more compact way
- [ ] Prevent rendering more than X characters wide (like default 80 for terminal width)
