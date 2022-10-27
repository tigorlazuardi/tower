.PHONY: install

GO_PACKAGES ?= $(shell go list ./... | grep -v 'mock')

doc: doc-binaries
	@./bin/go/git-chglog -o CHANGELOG.md
	@PYTHONPATH="$$(pwd)/bin/python:$$PYTHONPATH" ./bin/python/bin/markdown-pp README.mdpp -o README.md

doc-amend: doc
	@git add ./CHANGELOG.md || true
	@git add ./README.md || true
	@git add ./README.mdpp || true
	@git commit --amend --no-edit

test-binaries:
	@test -f "./bin/go/gotest" || GOBIN="$$(pwd)/bin/go" go install github.com/rakyll/gotest@v0.0.6

doc-binaries:
	@test -f "./bin/python/bin/markdown-pp" || pip install --target="$$(pwd)/bin/python" MarkdownPP
	@test -f "./bin/go/git-chglog" || GOBIN="$$(pwd)/bin/go" go install github.com/git-chglog/git-chglog/cmd/git-chglog@v0.15.1
	@test -d "./.chglog" || git-chglog --init

lint-binary:
	@test -f "./bin/go/golangci-lint" || GOBIN="$$(pwd)/bin/go" go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.50.0

lint: lint-binary
	@./bin/go/golangci-lint run

test: test-binaries
	@./bin/go/gotest -v ${GO_PACKAGES}

test-integration: test-binaries
	@IN_TEST=true ./bin/go/gotest -v ./...

git-hook:
	@test -f "./bin/go/lefthook" || GOBIN="$$(pwd)/bin/go" go install github.com/evilmartians/lefthook@v1.1.3
	@./bin/go/lefthook install


