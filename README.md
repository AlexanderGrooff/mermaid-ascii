# Mermaid ASCII

Render mermaid diagrams in your terminal:
```bash
$ cat test.mermaid
A --> C
A --> D
B --> C
C --> D
B --> D
$ mermaid --file test.mermaid
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

```
