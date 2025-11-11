default: help

help: ## list makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'


.PHONY: help fmt lint test mocks test_coverage

GO_PKGS   := $(shell go list -f {{.Dir}} ./...)

fmt: ## gofmt all files
	@go list -f {{.Dir}} ./... | xargs -I{} gofmt -w -s {}

lint: ## run golangci-lint
	@golangci-lint run

test: ## run tests
	@go test -race -v $(GO_FLAGS) -count=1 $(GO_PKGS)

test_coverage: ## run tests with coverage
	@go test -race -v $(GO_FLAGS) -count=1 -coverprofile=coverage.out -covermode=atomic $(GO_PKGS)
	@go tool cover -html coverage.out

mocks: ## generate mocks
	@go generate ./...
