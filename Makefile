export GO111MODULE=on

SHELL = /bin/bash -o pipefail

SRC = $(shell find . -name '*.go')

define PRE_COMMIT_HOOK
#!/bin/sh -xe
make lint
endef

install_hook := $(shell \
    if [ -z "$$INSTALL_HOOK" ]; then \
        INSTALL_HOOK=1 make .git/hooks/pre-commit; \
    fi; \
)

.PHONY: all
all: bin/astro bin/tvm

export PRE_COMMIT_HOOK
.git/hooks/pre-commit: Makefile
	echo "$$PRE_COMMIT_HOOK" >.git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit

bin:
	mkdir -p bin

bin/astro: bin $(SRC)
	go build -o bin/astro github.com/uber/astro/astro/cli/astro

bin/tvm: bin $(SRC)
	go build -o bin/tvm github.com/uber/astro/astro/tvm/cli/tvm

.PHONY: clean
clean:
	rm -f bin/astro
	rm -f bin/tvm

.PHONY: install
install:
	go install github.com/uber/astro/astro/cli/astro
	go install github.com/uber/astro/astro/tvm/cli/tvm

.PHONY: lint
lint:
	@f="$$(find . -name '*.go' ! -path './vendor/*' | xargs grep -L 'Licensed under the Apache License')"; \
	if [ ! -z "$$f" ]; then \
		echo "ERROR: Files missing license header:"$$'\n'"$$f" >&2; \
		exit 1; \
	fi;

.PHONY: release
release:
ifeq (, $(shell which zip))
	$(error "goreleaser not found. Follow https://goreleaser.com/install/ to install it")
endif
ifndef VERSION
	$(error "VERSION is not set, run `make release VERSION=1.2.3`")
endif
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)
	goreleaser release --rm-dist

.PHONY: test
test:
	go test -timeout 1m -coverprofile=.coverage.out ./... \
		|grep -v -E '^\?'
