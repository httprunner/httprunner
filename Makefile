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
	# =================== 参数说明 ===================
	# CGO_ENABLED=0         : 完全禁用 CGO，强制使用纯 Go 实现
	# -tags netgo,osusergo  : 使用 Go 的 net 和 user 包的纯 Go 实现，不依赖系统库
	# -trimpath             : 从二进制文件中删除所有文件系统路径，增加安全性和可重现性
	# -ldflags "-s -w"      :
	#   -s                  : 忽略符号表和调试信息
	#   -w                  : 忽略 DWARF 调试信息
	# -extldflags "-static" : 传递给外部链接器的标志，强制静态链接
	@echo "[info] build hrp cli tool"
	go mod tidy
	GOOS=${TARGET_OS} GOARCH=${TARGET_ARCH} CGO_ENABLED=0 go build -tags netgo,osusergo -trimpath -ldflags "\
		-s -w \
		-X 'github.com/httprunner/httprunner/v5/internal/version.GitCommit=$(shell git rev-parse HEAD)' \
		-X 'github.com/httprunner/httprunner/v5/internal/version.GitBranch=$(shell git rev-parse --abbrev-ref HEAD)' \
		-X 'github.com/httprunner/httprunner/v5/internal/version.GitAuthor=$(shell git log -1 --pretty=format:%an)' \
		-X 'github.com/httprunner/httprunner/v5/internal/version.BuildTime=$(shell date "+%y%m%d%H%M")'" \
		-o output/hrp ./cmd/cli
	./output/hrp -v

.PHONY: build-windows
build-windows: ## build hrp cli tool for Windows amd64
	@echo "[info] build hrp cli tool for Windows amd64"
	go mod tidy
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -tags netgo,osusergo -trimpath -ldflags "\
		-s -w \
		-X 'github.com/httprunner/httprunner/v5/internal/version.GitCommit=$(shell git rev-parse HEAD)' \
		-X 'github.com/httprunner/httprunner/v5/internal/version.GitBranch=$(shell git rev-parse --abbrev-ref HEAD)' \
		-X 'github.com/httprunner/httprunner/v5/internal/version.GitAuthor=$(shell git log -1 --pretty=format:%an)' \
		-X 'github.com/httprunner/httprunner/v5/internal/version.BuildTime=$(shell date "+%y%m%d%H%M")'" \
		-o output/hrp.exe ./cmd/cli
	@echo "[info] Windows binary built successfully: output/hrp.exe"

.PHONY: build-windows-arm64
build-windows-arm64: ## build hrp cli tool for Windows arm64
	@echo "[info] build hrp cli tool for Windows arm64"
	go mod tidy
	GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -tags netgo,osusergo -trimpath -ldflags "\
		-s -w \
		-X 'github.com/httprunner/httprunner/v5/internal/version.GitCommit=$(shell git rev-parse HEAD)' \
		-X 'github.com/httprunner/httprunner/v5/internal/version.GitBranch=$(shell git rev-parse --abbrev-ref HEAD)' \
		-X 'github.com/httprunner/httprunner/v5/internal/version.GitAuthor=$(shell git log -1 --pretty=format:%an)' \
		-X 'github.com/httprunner/httprunner/v5/internal/version.BuildTime=$(shell date "+%y%m%d%H%M")'" \
		-o output/hrp_arm64.exe ./cmd/cli
	@echo "[info] Windows ARM64 binary built successfully: output/hrp_arm64.exe"

.PHONY: build-linux
build-linux: ## build hrp cli tool for Linux amd64
	@echo "[info] build hrp cli tool for Linux amd64"
	go mod tidy
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags netgo,osusergo -trimpath -ldflags "\
		-s -w \
		-X 'github.com/httprunner/httprunner/v5/internal/version.GitCommit=$(shell git rev-parse HEAD)' \
		-X 'github.com/httprunner/httprunner/v5/internal/version.GitBranch=$(shell git rev-parse --abbrev-ref HEAD)' \
		-X 'github.com/httprunner/httprunner/v5/internal/version.GitAuthor=$(shell git log -1 --pretty=format:%an)' \
		-X 'github.com/httprunner/httprunner/v5/internal/version.BuildTime=$(shell date "+%y%m%d%H%M")'" \
		-o output/hrp_linux ./cmd/cli
	@echo "[info] Linux binary built successfully: output/hrp_linux"

.PHONY: build-all
build-all: build build-windows build-windows-arm64 build-linux ## build hrp cli tool for all platforms
	@echo "[info] All binaries built successfully"

.PHONY: install-hooks
install-hooks: ## install git hooks
	@find scripts -name "install-*-hook" | awk -F'-' '{s=$$2;for(i=3;i<NF;i++){s=s"-"$$i;}print s;}' | while read f; do bash "scripts/install-$$f-hook"; done
	@echo "[OK] install all hooks"

.PHONY: help
help: ## print make commands
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	  cut -d ":" -f1- | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
