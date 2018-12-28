SHELL = /bin/bash -o pipefail

export GO111MODULE=on

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

bin/astro: bin vendor
	go build -o bin/astro github.com/uber/astro/astro/cli/astro

bin/tvm: bin vendor
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

.PHONY: test
test: vendor
	go test -timeout 1m -coverprofile=.coverage.out ./... \
		|grep -v -E '^\?'

vendor:
	dep ensure
