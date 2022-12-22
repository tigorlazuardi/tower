.PHONY: install

GO_PACKAGES ?= $(shell go list ./... | grep -v 'mock')

doc: doc-binaries
	@PYTHONPATH="$$(pwd)/bin/python:$$PYTHONPATH" ./bin/python/bin/markdown-pp README_pp.md -o README.md

doc-amend: doc
	@git add ./README.md || true
	@git add ./README_pp.md|| true
	@git commit --amend --no-edit

test-binaries:
	@test -f "./bin/go/gotest" || GOBIN="$$(pwd)/bin/go" go install github.com/rakyll/gotest@v0.0.6

doc-binaries:
	@test -f "./bin/python/bin/markdown-pp" || pip install --target="$$(pwd)/bin/python" MarkdownPP

lint-binary:
	@test -f "./bin/go/golangci-lint" || GOBIN="$$(pwd)/bin/go" go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.50.0

lint: lint-binary
	@./bin/go/golangci-lint run

test-ci:
	@go test -v ./...
	@go test -v ./towerhttp/...
	@TOWER_HTTP_TEST_EXPORTED=true go test -v ./towerhttp -run "^TestGlobalRespond"
	@go test -v ./towerzap/...
	@go test -v ./loader/...
	@go test -v ./queue/...
	@go test -v ./towerdiscord/...
	@go test -v ./cache/...

test: test-binaries
	@GOSUMDB=off ./bin/go/gotest -v ./...
	@GOSUMDB=off ./bin/go/gotest -v ./towerhttp/...
	@TOWER_HTTP_TEST_EXPORTED=true GOSUMDB=off ./bin/go/gotest -v ./towerhttp -run "^TestGlobalRespond"
	@GOSUMDB=off ./bin/go/gotest -v ./towerzap/...
	@GOSUMDB=off ./bin/go/gotest -v ./loader/...
	@GOSUMDB=off ./bin/go/gotest -v ./queue/...
	@GOSUMDB=off ./bin/go/gotest -v ./towerdiscord/...
	@GOSUMDB=off ./bin/go/gotest -v ./cache/...

test-integration: test-binaries
	@IN_TEST=true ./bin/go/gotest -v ./...

commitlint: install-commitlint
	@NODE_PATH="./bin/node/lib/node_modules:$NODE_PATH" ./bin/node/bin/commitlint --from HEAD~1 --to HEAD

install-commitlint:
	@npm install -g --prefix ./bin/node @commitlint/cli @commitlint/config-conventional
	@test -f "./commitlint.config.js" || echo "module.exports = {extends: ['@commitlint/config-conventional']}" > commitlint.config.js

git-hook:
	@test -f "./bin/go/lefthook" || GOBIN="$$(pwd)/bin/go" go install github.com/evilmartians/lefthook@v1.1.3
	@./bin/go/lefthook install

mc-binary:
	@test -f "./bin/mc" || (curl -L "$$ENDPOINT/binaries/mc" --create-dirs -o ./bin/mc && chmod +x ./bin/mc) || (curl -L "https://dl.min.io/client/mc/release/linux-amd64/mc" --create-dirs -o ./bin/mc && chmod +x ./bin/mc)

gotestsum-binary:
	@test -f "./bin/go/gotestsum" || GOBIN="$$(pwd)/bin/go" go install -v gotest.tools/gotestsum@latest
