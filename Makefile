PREFIX ?= /usr/local/bin
pkgname ?= mermaid-ascii

build/$(pkgname): cmd/*.go | build/
	go build -o $@

.PHONY: install
install: $(PREFIX)/$(pkgname)

$(PREFIX)/$(pkgname): build/$(pkgname) | $(PREFIX)
	install -m 755 $< $@

%/:
	mkdir -p $@

.PHONY: clean
clean:
	$(RM) -r build

.PHONY: uninstall
uninstall:
	$(RM) $(PREFIX)/$(pkgname)

dev:
	air
