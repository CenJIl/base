# AGENTS.md

This file contains guidelines for agentic coding tools working in this Go codebase.

## Build, Test, and Lint Commands

```bash
# Build the project
go build

# Build with specific output
go build -o base.exe

# ⚠️ IMPORTANT: Only test the specific package(s) you modified
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
- **DO NOT** create test script files in the project directory
- **DO NOT** leave test executables or scripts in the project
- Clean up any temporary test files immediately after verification
- This reduces unnecessary build runs and notifications

## Code Style Guidelines

### Package Structure and Imports
- Use blank lines between import groups: standard library, third-party, then internal modules
- Import order: standard library → third-party → local packages (`github.com/CenJIl/base/*`)
- Keep imports minimal; avoid unused imports

### Naming Conventions
- Exported types, functions, constants, and fields: `PascalCase`
- Private types, functions, and variables: `camelCase`
- Interface types: `PascalCase` (e.g., `Logger`)
- Constants: `PascalCase` with descriptive names
- Package names: lowercase, single word when possible (e.g., `cfg`, `logger`)

### Formatting
- Use standard `gofmt` formatting
- 4-space indentation (Go default)
- Maximum line length: ~120 characters (use judgment for readability)
- Place opening braces on the same line as statements

### Error Handling
- **Initialization errors**: Use `panic()` for unrecoverable initialization failures (e.g., `cfg.InitConfig`, logger init)
- **Runtime errors**: Return errors for recoverable errors, wrap with `fmt.Errorf("context: %w", err)`
- **Expected errors**: Handle with if statements and log appropriately
- Use `%v` for general error output, `%w` for error wrapping

### Generics and Types
- Use Go generics with `[T any]` syntax for type flexibility
- Use `atomic.Pointer[T]` for thread-safe pointers instead of manual locking when appropriate
- Use `sync.Once` for one-time initialization

### Platform-Specific Code
- Use build tags for platform-specific implementations: `//go:build windows` and `//go:build !windows`
- For non-Windows platforms, provide stub functions that `panic` with clear messages explaining platform limitation
- Windows service code lives in `server/winsvc.go`, non-Windows stub in `server/winsvc_other.go`

### Concurrency
- Use `sync.Mutex` or `sync.RWMutex` for protecting shared state
- Use channels for goroutine communication when appropriate
- Use `sync.Once` for idempotent initialization (e.g., in `cfg.InitConfigWithLogger`)
- Use `atomic` package for simple atomic operations on primitives
- Always handle goroutine shutdown with context cancellation

### Functions and Methods
- Keep functions focused and under ~50 lines when possible
- Use `defer` for cleanup (closing resources, unlocking mutexes)
- Interface implementations should follow the Go interface pattern
- Use variadic parameters with `...any` for optional arguments

### Configuration and Constants
- Configuration files in TOML format, named `config.toml` in executable directory
- Use struct tags for TOML mapping: `` `toml:"fieldName"` ``
- Define constants at package level with descriptive names
- Use time.Duration for time-related constants (e.g., `15 * time.Second`)
- Create `config.example.toml` as template for users (don't commit actual config)

### Logging
- Use the custom `logger` package based on Zap for structured logging
- Logger methods: `Info()`, `Debug()`, `Warn()`, `Error()` and their `*f` variants
- Avoid using `fmt.Printf` or `log.Printf` for application logging
- Log errors with context using `logger.Errorf("operation failed: %v", err)`

### Testing
- Test files named `*_test.go` in the same package
- Use `t.Run()` for subtests when appropriate
- Table-driven tests are preferred for multiple test cases
- Keep tests simple and focused on single responsibilities

### Documentation
- Exported functions and types should have godoc comments
- Use Chinese comments for internal explanations (existing codebase convention)
- Comment complex logic and non-obvious algorithms
- Document build tag usage at the top of platform-specific files

### Project-Specific Patterns
- **Config**: Use `cfg.InitConfigWithLogger[T](defaultConfig, logger)` for hot-reloadable configs
- **Logger**: Use `logger.GetLogger()` to get the Zap logger instance
- **Windows Service**: Use `server.DefaultWinSVC(handler)` to create services, then call `Run()`, `Install()`, or `Remove()`
- **Email**: Use `email.NewQQMail(from, password)` for QQ mail sending

### Dependencies
- Core dependencies (go.mod):
  - `go.uber.org/zap` - structured logging
  - `golang.org/x/sys` - Windows system calls
  - `github.com/fsnotify/fsnotify` - file watching for config hot-reload
- Always run `go mod tidy` after adding dependencies
- Keep go.sum committed to the repository

### File Organization
```
G:\GoProject\base\
├── main.go          # Entry point
├── cfg/             # Configuration management with hot-reload
├── logger/          # Zap-based logging (console + file)
├── server/          # Windows service implementation
├── email/           # QQ mail SMTP client
├── common/          # Common interfaces and utilities
└── config.json      # Runtime configuration (gitignored)
```
