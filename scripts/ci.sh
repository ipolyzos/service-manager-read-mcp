#!/usr/bin/env bash
# Same checks as GitHub Actions (`.github/workflows/ci.yml`); run locally before push or release prep.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo "== 1/7 gofmt =="
out="$(gofmt -l .)"
if [[ -n "$out" ]]; then
	echo "Unformatted files — run: gofmt -w on the paths below" >&2
	echo "$out" >&2
	exit 1
fi

echo "== 2/7 go mod verify =="
go mod verify

echo "== 3/7 go mod tidy (working tree must stay clean) =="
go mod tidy
git diff --exit-code go.mod go.sum

echo "== 4/7 go vet =="
go vet ./...

echo "== 5/7 golangci-lint =="
if ! command -v golangci-lint >/dev/null 2>&1; then
	echo "Install golangci-lint: https://golangci-lint.run/welcome/install/" >&2
	exit 1
fi
golangci-lint run ./... --timeout=5m

echo "== 6/7 go build ./... =="
go build ./...

echo "== 7/7 tests: unit (-short) then stdio smoke =="
go test -count=1 -race -short ./...
go test -count=1 -race ./cmd/service-manager-read-mcp -run '^TestCLI_stdioSmoke$'

echo "All checks passed."
