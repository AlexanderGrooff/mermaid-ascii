PREFIX ?= /usr/local/bin
pkgname ?= mermaid-ascii

targets += $(PREFIX)/$(pkgname)

ifneq (,$(wildcard /usr/share/bash-completion/completions/))
  targets += /usr/share/bash-completion/completions/$(pkgname)
endif

all: build/$(pkgname) build/completions/bash build/completions/zsh build/completions/fish

build/$(pkgname): cmd/*.go | build/
	go build -o $@

.PHONY: install
install: $(targets)

$(PREFIX)/$(pkgname): build/$(pkgname) | $(PREFIX)
	install -m 755 $< $@
/usr/share/bash-completion/completions/$(pkgname): build/completions/bash
	install -m 755 $< $@

build/completions/%: build/$(pkgname) | build/completions/
	./$< completion $(@F) > $@

%/:
	mkdir -p $@

.PHONY: clean
clean:
	$(RM) -r build

.PHONY: uninstall
uninstall:
	$(RM) $(targets)

.PHONY: test
test:
	go test ./... -v

.PHONY: docker-build
docker-build:
	docker build -t mermaid-ascii:latest .

.PHONY: docker-test
docker-test:
	docker build --target test -t mermaid-ascii:test -f Dockerfile.test .

.PHONY: docker-run
docker-run:
	docker run -i mermaid-ascii:latest

dev:
	air
