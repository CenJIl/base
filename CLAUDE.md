# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build, Test, and Lint Commands

```bash
# Build the project
go build

# Build with specific output
go build -o base.exe

# ⚠️ CRITICAL: Only test the specific package(s) you modified
# DO NOT run all tests with `go test ./...`
# This reduces notifications and focuses on changes

# Run tests for a specific package (preferred)
go test ./web
go test ./server
go test ./logger
go test ./cfg
go test ./email

# Run a single test function
go test ./server -run TestDefault
go test ./server -run TestSpecificFunction

# Run tests with verbose output (only for specific package)
go test -v ./web

# Run tests with coverage (only for specific package)
go test -cover ./web
go test -coverprofile=coverage.out ./web

# Install linter (if not already installed)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter (only on modified packages)
golangci-lint run ./web

# Format code
go fmt ./...
```

### Testing Rules
- **NEVER** run `go test ./...` after making changes
- **ONLY** run tests for the specific package(s) you modified
- **FOCUSED TESTING**: Test only the module you changed
  - Modified `web/`? Run: `go test ./web`
  - Modified `logger/`? Run: `go test ./logger`
  - Modified `cfg/`? Run: `go test ./cfg`
- **DO NOT** create test script files in the project directory
- **DO NOT** leave test executables or scripts in the project
- Clean up any temporary test files immediately after verification
- This reduces unnecessary build runs and notifications

## Project Nature

This is a **utility helper library + web scaffolding repository**, NOT a standalone application.

**Critical Rules**:
- ❌ **NO main.go files** - This is a tool library, all demos/examples must be in `*_test.go` files
- ✅ **ONLY test-based examples** - All usage examples must be in module-specific `*_test.go` files
- ❌ **DO NOT create documentation files** - Use code comments instead
- ❌ **DO NOT create markdown guides** - Add comments directly in code
- ❌ **DO NOT create README files** - Add comments directly in code
- ❌ **DO NOT create .md files** - All documentation must be in code comments
- ✅ **FOCUS on code annotations** - Document usage through Go comments and examples
- **TEST what you modify** - Only run tests for specific packages you change
- ❌ **NO random documentation** - This is a scaffolding library, not a tutorial site

When adding features, document them directly in the code with clear comments, not separate markdown files.

**Example Structure**:
- `web/demo_test.go` - Contains `ExampleScenario1_FileUpload()`, `ExampleScenario2_APIErrorHandling()`, etc.
- Each test function demonstrates a complete usage scenario
- Run with: `go test -v ./web -run ExampleScenario1`

### Core Packages

- **cfg** - Configuration management with hot-reload support using TOML files
- **logger** - Zap-based structured logging with console and file output
- **web** - Gin web framework wrapper with middleware (recovery, i18n, logger, response)
- **server** - Windows service implementation with install/remove capabilities
- **email** - QQ mail SMTP client for sending emails
- **common** - Shared interfaces (Logger) and utilities

### Architecture Patterns

**Configuration System**: Uses Go generics with `cfg.InitConfig[T](defaultConfig)` where T is your config struct. The system:
- Looks for `config.toml` in the executable directory
- Creates the file with defaults if missing
- Watches for file changes and hot-reloads via fsnotify
- Provides `cfg.GetCfg[T]()` to access current config
- Supports change callbacks via `cfg.OnConfigChange[T](func(*T))`

**Logging System**: Zap-based structured logging with:
- Console output (always enabled) with colored level tags
- File output (only when running as Windows service) to `logs/app.log`
- Dynamic log level control via `logger.UpdateLogLevel(level string)`
- Automatic log rotation (20MB max, 10 backups, 30 days retention)
- Platform-specific: `zap_windows.go` for Windows service detection, `zap_other.go` stub

**Web Framework**: Gin wrapper with pre-configured middleware stack:
- Recovery middleware for panic handling
- i18n middleware with TOML locale files
- Request/response logging middleware
- Unified JSON response format with trace IDs
- Must embed `web.WebBaseConfig` in your config struct

**Windows Service**: Full service lifecycle management:
- Service name auto-extracted from handler function name
- Install/remove with admin elevation
- Graceful shutdown with configurable timeout (default 15s)
- Automatic working directory switching to executable location
- Platform-specific: `winsvc.go` for Windows, `winsvc_other.go` stub (panics)

### Key Conventions

**Initialization Order** (main.go pattern):
1. Define config struct with embedded `web.WebBaseConfig`
2. Call `cfg.InitConfig[YourConfig](defaultConfigTOML)`
3. Get config via `cfg.GetCfg[YourConfig]()`
4. Initialize web engine with `web.NewGin(config.Web)`
5. Start server

**Config File**: `config.toml` (gitignored) in executable directory
- TOML format with struct tags: `` `toml:"fieldName"` ``
- Example template should be provided as `config.example.toml`

**Platform-Specific Code**: Use build tags `//go:build windows` and `//go:build !windows`
- Non-Windows platforms provide stub functions that panic with clear messages
- Windows implementations live in `*_windows.go` files
- Non-Windows stubs live in `*_other.go` files

**Error Handling**:
- Initialization failures: use `panic()` (unrecoverable)
- Runtime errors: return wrapped errors with `fmt.Errorf("context: %w", err)`

**Concurrency Patterns**:
- `sync.Once` for one-time initialization (cfg init, web middleware)
- `atomic.Pointer[T]` for thread-safe config storage
- `context.Context` for graceful shutdown signaling
- Channels for goroutine communication

### Module Path
All imports use `github.com/CenJIl/base/*` for local packages.

### Gitignored Items
- `config.toml` - runtime configuration
- `logs/` - log file output directory
- `*.exe`, `*.test` - build artifacts
- `coverage.*` - test coverage reports
- `.env` - environment variables
