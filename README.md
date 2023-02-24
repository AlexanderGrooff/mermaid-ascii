# Mermaid ASCII

Render mermaid diagrams in your terminal:
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
