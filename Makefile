.PHONY: all
all: bin/astro bin/tvm

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

.PHONY: test
test: vendor
	go test -timeout 1m -coverprofile=.coverage.out ./... \
		|grep -v -E '^\?'

vendor:
	dep ensure
