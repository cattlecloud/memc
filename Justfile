set shell := ["bash", "-c"]

# tidy, vet, lint, and test the source tree
default: tidy vet lint test

# run tests
test:
    go test -count=1 -v -race ./...

# scan for missing copywrite headers
copywrite:
    copywrite \
        --config .github/workflows/scripts/copywrite.hcl headers \
        --spdx "BSD-3-Clause"

# lint source with golangci-lint
lint:
    golangci-lint run \
        --config .github/workflows/scripts/golangci.yaml

# scan source with go ver
vet:
    go vet ./...

# tidy go modules
tidy:
    go mod tidy
