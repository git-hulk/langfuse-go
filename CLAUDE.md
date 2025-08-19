# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

This is the **langfuse-go** client library for interacting with the Langfuse platform. It provides Go bindings for tracing, prompt management, model management, scores, datasets, sessions, LLM connections, organizations, comments, and annotations functionality.

## Key Architecture

The codebase follows a modular structure with clear separation of concerns:

- **Main Client (`langfuse.go`)**: Entry point that orchestrates all functionality via composition
- **Package-based modules**: Each feature area is in its own package under `pkg/` with consistent patterns
- **Generic Batch Processing**: `pkg/batch/` provides a type-safe generic batch processor for efficient API ingestion
- **HTTP Client**: Uses `go-resty/resty` with `/api/public` base URL and basic auth configured once
- **Shared Components**: `pkg/common/` contains shared HTTP utilities and model units

### Core Components

1. **LangFuse Client**: Main struct that composes all feature clients, initialized with host + credentials
2. **Traces & Observations**: Hierarchical tracing system with traces containing observations (spans), using batch ingestion
3. **Generic Batch Processor**: Type-safe buffering system with configurable batch sizes, flush intervals, and worker pools
4. **Feature Clients**: Independent clients for each API area (prompts, models, scores, etc.) sharing HTTP configuration
5. **Annotations System**: Queue-based annotation workflows with items and assignments
6. **Comments System**: Contextual comments for traces, observations, and sessions

### Client Architecture Pattern

All feature clients follow the same pattern:
- Accept a configured `*resty.Client` in constructor
- Provide CRUD operations with context support
- Use consistent naming: `Get()`, `List()`, `Create()`, `Delete()`
- Return structured responses with proper error handling
- Use validation on request structs

### Batch Processing Architecture

The traces system uses a sophisticated generic batch processor (`pkg/batch/Processor[T]`) that:
- Buffers incoming records in a channel-based queue
- Batches records by size (default 32) or time interval (default 3s)
- Uses configurable worker goroutines for parallel processing
- Provides graceful shutdown with timeout handling
- Supports any type implementing the `Sender[T]` interface

## Development Commands

### Testing
```bash
make test                    # Run all tests with race detector (-race -count=1)
go test ./...               # Standard Go test runner  
go test ./pkg/annotations/  # Test specific package
go test -v ./pkg/traces/    # Verbose output for specific package
go run integration/integration.go  # Run integration tests (requires env setup)
```

### Code Formatting
```bash
make format                 # Format with goimports + gofmt (includes local import ordering)
goimports -w -local github.com/git-hulk/langfuse-go ./...
```

### Build & Linting
```bash
go build ./...              # Build all packages
golangci-lint run           # Lint (CI uses v1.64.7)
```

## API Naming Conventions

### Parameter and Path Naming
- Use proper Go casing for struct fields: `QueueID`, `ItemID`, `ConfigID`
- Keep JSON tags in camelCase for API compatibility: `json:"queueId"`, `json:"itemId"`
- Use proper casing in path parameters: `SetPathParam("queueID", ...)` and `{queueID}` in URLs
- Error messages should use proper casing: `"'queueID' is required"`

### URL Structure
- API paths omit `/api/public` prefix (set in client base URL)
- Use pattern: `/resource-name/{resourceID}/sub-resource/{subResourceID}`
- Example: `/annotation-queues/{queueID}/items/{itemID}`

## Code Patterns

### Client Initialization
```go
// Main client sets base URL and auth once
restyCli := resty.New().
    SetBaseURL(host+"/api/public").
    SetBasicAuth(publicKey, secretKey)

// Feature clients reuse the configured HTTP client
client := features.NewClient(restyCli)
```

### Error Handling & Validation
- All request structs implement `validate() error` methods
- Use proper Go error wrapping: `fmt.Errorf("failed to X: %w", err)`
- Check required fields and return descriptive errors
- HTTP errors include status codes and response bodies

### Testing Patterns
- Use `github.com/stretchr/testify` for assertions (`require`, `assert`)
- Table-driven tests with `name`, `input`, `expected`, `wantErr` fields
- Use `httptest.NewServer()` for HTTP client testing
- Test both success and error cases including validation failures

### JSON Struct Tags
- Use `omitempty` for optional fields: `json:"field,omitempty"`
- Match API field naming exactly in JSON tags
- Use pointer types for truly optional fields that can be nil

### Function and Method Patterns
- Include context parameters for functions that perform I/O operations
- Add proper documentation comments for exported functions
- Always handle errors explicitly, never ignore them
- Use error wrapping to provide context: `fmt.Errorf("failed to X: %w", err)`
- Start error messages with lowercase letters

### Struct and Interface Design
- Generate structs with proper field tags for JSON
- Use `any` instead of `interface{}` for generic types in Go 1.18+
- Keep interfaces small and focused (Interface Segregation Principle)
- Use descriptive names ending with "-er" when appropriate
- Use pointer types for optional fields that can be nil

### Constants and Naming
- Use `Enabled` suffix for feature toggles: `CacheEnabled`, `LoggingEnabled`
- Use `Is` prefix for state checks: `IsActive`, `IsValid`
- Use ALL_CAPS with underscores for package-level constants
- Group related constants in blocks
