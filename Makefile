SHELL=/usr/bin/env bash

MACOS_MIN   := 11.0
TARGET_OS   := darwin
TARGET_ARCH := amd64

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
	go mod tidy
	GOOS=${TARGET_OS} GOARCH=${TARGET_ARCH} go build -ldflags "\
		-s -w \
		-X 'github.com/httprunner/httprunner/v5/internal/version.GitCommit=$(shell git rev-parse HEAD)' \
		-X 'github.com/httprunner/httprunner/v5/internal/version.GitBranch=$(shell git rev-parse --abbrev-ref HEAD)' \
		-X 'github.com/httprunner/httprunner/v5/internal/version.GitAuthor=$(shell git log -1 --pretty=format:%an)' \
		-X 'github.com/httprunner/httprunner/v5/internal/version.BuildTime=$(shell date "+%y%m%d%H%M")'" \
		-o output/hrp ./cmd/cli
	./output/hrp -v

.PHONY: install-hooks
install-hooks: ## install git hooks
	@find scripts -name "install-*-hook" | awk -F'-' '{s=$$2;for(i=3;i<NF;i++){s=s"-"$$i;}print s;}' | while read f; do bash "scripts/install-$$f-hook"; done
	@echo "[OK] install all hooks"

.PHONY: help
help: ## print make commands
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	  cut -d ":" -f1- | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
