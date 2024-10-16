set shell := ["bash", "-c"]

# vet, lint, and test the source tree
default: vet lint test

# run go test on source tree
test:
    go test -count=1 -v -race ./...

# apply copywrite headers to source tree
copywrite:
    copywrite \
        --config .github/workflows/scripts/copywrite.hcl headers \
        --spdx "BSD-3-Clause"

# apply golang linter to source tree
lint:
    golangci-lint run \
        --config .github/workflows/scripts/golangci.yaml

# run go vet on source tree
vet:
    go vet ./...
