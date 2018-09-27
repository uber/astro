#!/bin/bash
exec go run $GOPATH/src/github.com/uber/astro/astro/tvm/cli/tvm/main.go "$@"
