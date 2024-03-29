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

Or using Nix:
```bash
$ git clone
$ cd mermaid-ascii
$ nix build
$ ./result/bin/mermaid-ascii --help
```

## Usage
```bash
$ cat test.mermaid
graph LR
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
graph LR
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

# Top-down layout
$ cat test.mermaid
graph TD
A --> B
B --> C
B --> D
C --> D
X --> C
$ mermaid-ascii -f ./test.mermaid -y 5
+---+          +---+
|   |          |   |
| A |          | X |
|   |          |   |
+---+          +---+
  |              |
  |              |
  |              |
  |              |
  v              |
+---+            /
|   |           /
| B |          /
|   |         /
+---+        /
  \         /
  |\       /
  | \     /
  |  \   /
  v   \ /
+---+  /       +---+
|   | / \      |   |
| C |<-------->| D |
|   |          |   |
+---+          +---+

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

Colored output is also supported (given that your terminal supports it) using the `classDef` syntax:

```bash
graph LR
classDef example1 color:#ff0000
classDef example2 color:#00ff00
classDef example3 color:#0000ff
test1:::example1 --> test2
test2:::example2 --> test3:::example3
```

This results in the following graph:

![](docs/colored_graph.png)

## TODOs

The baseline components for Mermaid work, but there are a lot of things that are not supported yet. Here's a list of things that are not yet supported:

### Syntax support

- [x] Labelled edges (like `A -->|label| B`)
- [x] Graph directions like `graph LR` and `graph TB`
- [x] `classDef` and `class`
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
