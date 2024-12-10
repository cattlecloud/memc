set shell := ["bash", "-u", "-c"]

[private]
default:
    @just --list

# run tests over source tree
test:
    go test -count=1 -v -race ./...

# scan for missing copywrite headers
copywrite:
    copywrite \
        --config .github/workflows/scripts/copywrite.hcl headers \
        --spdx "BSD-3-Clause"

# run lint over source tree
lint: vet
    golangci-lint run \
        --config .github/workflows/scripts/golangci.yaml

# run vet over source tree
vet:
    go vet ./...

# tidy go modules
tidy:
    go mod tidy
