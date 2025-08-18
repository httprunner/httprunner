# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

HttpRunner v5 is a comprehensive testing framework written in Go that supports API testing, load testing, and UI automation across multiple platforms (Android/iOS/Harmony/Browser). The framework integrates LLM technology for intelligent test automation and uses a pure visual-driven approach (OCR/CV/VLM) for UI testing.

## Development Commands

### Building
- `make build` - Build the hrp CLI tool with static linking and embedded version info
- `go build -o output/hrp ./cmd/cli` - Alternative build command
- `make test` - Run unit tests with race detection

### Testing
- `go test -race -v ./...` - Run all tests with race detection
- `go test -v ./tests/...` - Run test suite only
- `go test -v ./uixt/...` - Run UI automation tests
- `go test -v ./cmd/...` - Run CLI command tests

### Code Quality
- `go mod tidy` - Clean up dependencies
- `gofmt -w .` - Format code
- Pre-commit hooks are available in `scripts/` directory

## Core Architecture

### Main Components

**Core Testing Engine**
- `runner.go` - Main test runner (HRPRunner, CaseRunner, SessionRunner)
- `testcase.go` - Test case definitions and loading (ITestCase interface)
- `step.go` - Step definitions and configurations
- `step_*.go` - Specific step implementations (request, api, testcase, ui, etc.)

**Step Types**
- `step_request.go` - HTTP/HTTPS requests
- `step_api.go` - API calls with parameters
- `step_testcase.go` - Nested test cases
- `step_websocket.go` - WebSocket communication
- `step_ui.go` - UI automation steps
- `step_transaction.go` - Transaction grouping
- `step_rendezvous.go` - Synchronization points
- `step_shell.go` - Shell command execution
- `step_function.go` - Custom function calls

**UI Automation (uixt/)**
- `device.go` - Device abstraction interface (IDevice)
- `driver.go` - Driver interface and session management
- `android_*.go` - Android platform implementation (ADB/UIAutomator2)
- `ios_*.go` - iOS platform implementation (WDA)
- `harmony_*.go` - HarmonyOS implementation (HDC)
- `browser_*.go` - Web browser automation
- `ai/` - AI-powered UI interaction (OCR/VLM)

**CLI Interface (cmd/)**
- `root.go` - Root command and global configuration
- `run.go` - Test execution
- `server.go` - HTTP server mode
- `convert.go` - Format conversion utilities
- `build.go` - Plugin building
- `adb/` - Android device management
- `ios/` - iOS device management

### Plugin System

The framework supports both Go and Python plugins:
- `build.go` - Plugin compilation system
- `plugin.go` - Plugin interface definitions
- Templates in `internal/scaffold/templates/plugin/`

### Configuration Management

- `config.go` - Global configuration
- `internal/config/` - Environment and settings management
- Environment variables and .env file support

## Key Design Patterns

### Interface-Driven Architecture
- `ITestCase` interface for different test case sources
- `IDevice` interface for multi-platform support
- `IDriver` interface for different automation drivers

### Step-Based Testing
- Each test consists of configurable steps
- Steps support setup/teardown hooks
- Variables and parameters flow between steps

### Plugin Architecture
- Hashicorp go-plugin for Go plugins
- Python plugin support via funplugin
- Template-based plugin generation

## Testing Approach

### Test Formats Supported
- YAML/JSON test cases
- Go test files
- Python pytest integration
- HAR, Postman, cURL conversion

### UI Testing Strategy
- Pure visual-driven (no element locators)
- OCR/VLM for text recognition
- Cross-platform unified API
- AI-powered interaction planning

## Development Guidelines

### Code Structure
- Core framework logic in root directory
- Platform-specific implementations in `uixt/`
- CLI commands in `cmd/`
- Internal utilities in `internal/`
- Examples in `examples/`

### Code Standards
- All code comments must be written in English
- All documentation must be written in Chinese

### Dependencies
- Go 1.23+ required
- Uses Cobra for CLI
- Integrates with multiple automation frameworks
- LLM integration via CloudWeGo Eino

### Build Configuration
- Static linking for deployment
- Version info embedded via ldflags
- Cross-platform builds supported
