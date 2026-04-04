.PHONY: ci build build-all test test-short test-smoke vet fmt-check mod-verify

# Full local check suite (matches CI intent; requires golangci-lint on PATH).
ci:
	./scripts/ci.sh

# Installable binary under bin/ (gitignored).
build:
	mkdir -p bin
	go build -o bin/service-manager-read-mcp ./cmd/service-manager-read-mcp

# Compile all packages (matches CI `go build ./...`).
build-all:
	go build ./...

# Race + short: skips TestCLI_stdioSmoke in cmd/service-manager-read-mcp.
test-short:
	go test -count=1 -race -short ./...

# Subprocess stdio smoke (not covered by -short).
test-smoke:
	go test -count=1 -race ./cmd/service-manager-read-mcp -run '^TestCLI_stdioSmoke$$'

vet:
	go vet ./...

fmt-check:
	@test -z "$$(gofmt -l .)" || (echo "run gofmt -w on:"; gofmt -l .; exit 1)

mod-verify:
	go mod verify
