.PHONY: build build-release

VERSION_PKG := github.com/ipolyzos/service-manager-read-mcp/internal/version
# Optional release semver, e.g. VERSION=0.4.0 make build-release
VERSION ?=
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
RELEASE_LDFLAGS := -X '$(VERSION_PKG).Version=$(VERSION)' -X '$(VERSION_PKG).Commit=$(COMMIT)' -X '$(VERSION_PKG).BuildDate=$(BUILD_DATE)'

# Default developer build (no embedded release metadata).
build:
	go build -o bin/service-manager-read-mcp ./cmd/service-manager-read-mcp

# Release-style binary: -trimpath and Version/Commit/BuildDate when set.
build-release:
	mkdir -p bin
	go build -trimpath -ldflags "$(RELEASE_LDFLAGS)" -o bin/service-manager-read-mcp ./cmd/service-manager-read-mcp
