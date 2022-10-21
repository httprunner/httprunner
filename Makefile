SHELL=/usr/bin/env bash

.DEFAULT_GOAL=help

.PHONY: test
test: ## run unit tests
	@echo "[info] run unit tests"
	@echo "go test -race -v ./..."
	@go test -race -v ./...

.PHONY: bump
bump: ## bump hrp version, e.g. make bump version=4.0.0
	@echo "[info] bump hrp version"
	@. scripts/bump_version.sh $(version)

.PHONY: build
build: ## build hrp cli tool
	@echo "[info] build hrp cli tool"
	@. scripts/build.sh $(tags)

.PHONY: install-hooks
install-hooks: ## install git hooks
	@find scripts -name "install-*-hook" | awk -F'-' '{s=$$2;for(i=3;i<NF;i++){s=s"-"$$i;}print s;}' | while read f; do bash "scripts/install-$$f-hook"; done
	@echo "[OK] install all hooks"

.PHONY: help
help: ## print make commands
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	  cut -d ":" -f1- | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
