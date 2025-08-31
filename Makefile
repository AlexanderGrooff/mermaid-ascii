PREFIX ?= /usr/local/bin

mermaid-ascii: cmd/*.go
	go build

.PHONY: install
install: $(PREFIX)/mermaid-ascii

$(PREFIX)/mermaid-ascii: mermaid-ascii | $(PREFIX)
	install -m 755 $< $@

.PHONY: clean
clean:
	$(RM) mermaid-ascii

.PHONY: uninstall
uninstall:
	$(RM) $(PREFIX)/mermaid-ascii
