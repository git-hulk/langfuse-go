SHELL := /bin/bash
GO ?= go
GOIMPORTS ?= goimports
PKGS := $(shell $(GO) list ./... | grep -v /vendor/)
GO_FILES := $(shell find . -type f -name '*.go' -not -path './vendor/*')

.PHONY: test format

## Run all tests with race detector enabled
test:
	@echo '==> Running tests (race mode)'
	@$(GO) test -race -count=1 $(PKGS)

## Format code using gofmt
format:
	@echo '==> Running goimports'
	@$(GOIMPORTS) -w -local github.com/git-hulk/langfuse-go $(GO_FILES)
	@echo '==> Formatting Go files'
	@gofmt -l -w $(GO_FILES)
