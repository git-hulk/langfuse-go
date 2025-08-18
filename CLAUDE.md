# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

This is the **langfuse-go** client library for interacting with the Langfuse platform. It provides Go bindings for tracing, prompt management, model management, and comments functionality.

## Key Architecture

The codebase follows a modular structure:

- **Main Client (`langfuse.go`)**: Entry point that orchestrates all functionality via composition
- **Package-based modules**: Each feature (traces, prompts, models, comments) is in its own package under `pkg/`
- **Batch Processing**: Generic batch processor (`pkg/batch/`) handles efficient data ingestion
- **HTTP Client**: Uses `go-resty/resty` for all API communication with basic auth

### Core Components

1. **LangFuse Client**: Main struct that composes all feature clients
2. **Traces**: Hierarchical tracing with traces containing observations (spans)
3. **Batch Processor**: Generic buffering system for efficient API calls
4. **Feature Clients**: Independent clients for prompts, models, and comments

## Development Commands

### Testing
```bash
make test          # Run all tests with race detector
go test ./...      # Standard Go test runner
go test -race ./...  # With race detector manually
```

### Code Formatting
```bash
make format        # Format code with goimports and gofmt
```

### Build
```bash
go build ./...     # Build all packages
```

### Linting
The CI uses `golangci-lint` v1.64.7. Install and run locally:
```bash
golangci-lint run
```

## Test Structure

- Tests are co-located with source files using `_test.go` suffix
- Uses `testify` for assertions (`github.com/stretchr/testify`)
- All new functionality requires unit tests

## Dependencies

- **HTTP Client**: `github.com/go-resty/resty/v2` for API communication
- **UUID Generation**: `github.com/gofrs/uuid/v5`
- **Collections**: `github.com/hashicorp/go-set/v3`
- **Testing**: `github.com/stretchr/testify`

## Code Patterns

### Client Initialization
All feature clients are initialized through the main `LangFuse` struct with shared HTTP client configuration.

### Batch Processing
The traces functionality uses a generic batch processor that buffers records and sends them in configurable batches with flush intervals.

### Error Handling
Standard Go error handling patterns. API errors are wrapped and returned up the call stack.

### Struct Composition
The main client uses composition rather than inheritance, with each feature area having its own client struct.
