# Project Overview

This project is a Go client library for the LangFuse API. It provides a simple way to instrument Go applications to send tracing data to LangFuse. The library supports creating traces and spans, and it batches the data to be sent to the LangFuse API efficiently.

**Key Technologies:**

*   Go
*   [resty](https://github.com/go-resty/resty) for HTTP requests

**Architecture:**

The library is composed of three main parts:

1.  **`LangFuse` client:** The main entry point for interacting with the library. It's used to create traces.
2.  **`Trace` and `Observation` (Span):** These are the data structures that represent the tracing data.
3.  **Batch Processor:** A generic batch processor that collects traces and sends them to the LangFuse API in the background.

# Building and Running

## Building the library

To build the library, you can use the standard Go build command:

```bash
go build ./...
```

## Running tests

The project uses the standard Go testing framework. To run the tests, use the following command:

```bash
go test ./...
```

# Development Conventions

## Code Style

The project follows the standard Go code style. It's recommended to use `gofmt` to format the code before committing.

## Testing

All new features should be accompanied by unit tests. The tests are located in the same package as the code they are testing, with the `_test.go` suffix.
