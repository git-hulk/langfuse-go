# Repository Guidelines

## Project Structure & Module Organization
The module `github.com/git-hulk/langfuse-go` targets Go 1.23. High-level client code lives in `langfuse.go`, while domain-specific APIs sit under `pkg/<domain>` (for example `pkg/traces`, `pkg/datasets`, `pkg/prompts`). Each package contains request/response structs plus helpers, and keeps its tests alongside implementation files. Integration examples and fixtures are under `integration/`, with `integration/testdata/` holding sample payloads. Shared configs such as `openapi.yml` and `Makefile` stay in the root.

## Build, Test, and Development Commands
Use `go build ./...` to ensure all packages compile. `make test` runs `go test -race -count=1 ./...` against every package—prefer this before opening a PR. Targeted suites (e.g., `go test ./pkg/traces`) keep iteration fast. Format code with `make format`, which chains `goimports -local github.com/git-hulk/langfuse-go` and `gofmt` to enforce import grouping and canonical spacing.

## Coding Style & Naming Conventions
Stick to idiomatic Go: tabs for indentation, exported APIs prefixed with capital letters, and constructors aliased as `New<Type>()`. When defining structs or params, mirror Langfuse field names (e.g., `promptName`, `promptVersion`) and keep JSON tags in snake_case to match the API. Avoid introducing custom logger abstractions—use the existing `pkg/logger` utilities. Always run `goimports` to maintain grouped imports and remove unused symbols.

## Testing Guidelines
Unit tests use the standard library plus `testify` assertions. Name tests `Test<Resource><Behavior>` and table tests `Test<Resource>_<Case>`. For race-sensitive code, rely on `make test`; for coverage, `go test -cover ./pkg/...` is acceptable in CI logs. When adding new API clients, include integration-style tests that hit the mock fixtures in `integration/testdata/` so serialization stays stable.

## Commit & Pull Request Guidelines
Recent history uses short, imperative commit messages with optional PR tags (e.g., `Add support for promptName (#10)`). Follow that style and keep one logical change per commit. PRs should describe API changes, list new commands or env vars, and mention related Langfuse issues. Always include the command output for `make test`, note any breaking changes, and attach screenshots only when UI behavior (if any) is affected.
