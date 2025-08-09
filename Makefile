.PHONY: build test lint tidy ci help

BIN := ./bin
PKG := ./...
MAIN := ./cmd/ancli

help: ## Show targets
	@grep -E '^[a-zA-Z_-]+:.*?##' $(MAKEFILE_LIST) | \
		awk -F':.*?## ' '{printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

build: ## Compile ancli
	go build -o $(BIN)/ancli $(MAIN)

test: ## Run unit tests
	go test -race $(PKG)

lint: ## Run golangci-lint
	golangci-lint run

tidy: ## Tidy go.mod/sum
	go mod tidy
	go mod verify

ci: tidy lint test build ## Compose target for CI
