SHELL=/usr/bin/env bash

.DEFAULT_GOAL=help

.PHONY: test
test: ## run unit tests
	@echo "[info] run unit tests"
	@echo "go test -race -v ./..."
	@go test -race -v ./...

.PHONY: bump
bump: ## bump hrp version
	@echo "[info] bump hrp version"
	@. cli/scripts/bump_version.sh $(version)

.PHONY: build
build: ## build hrp cli tool
	@echo "[info] build hrp cli tool"
	@. cli/scripts/build.sh

.PHONY: help
help: ## print make commands
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	  cut -d ":" -f1- | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
