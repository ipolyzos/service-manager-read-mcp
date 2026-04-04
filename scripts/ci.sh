#!/usr/bin/env bash
# Same checks as GitHub Actions (`.github/workflows/ci-verify.yml`, invoked from
# `ci.yml` and `release.yml`). Run locally before push or release prep.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

GOLANGCI_PIN="$(tr -d '[:space:]' < "$ROOT/scripts/golangci-version")"

echo "== 1/8 Go toolchain (must match go.mod) =="
want_go="$(go list -m -f '{{.GoVersion}}' . 2>/dev/null || true)"
if [[ -z "$want_go" || "$want_go" == "<no value>" ]]; then
	echo "Could not read Go version from the module (go list -m -f '{{.GoVersion}}')." >&2
	exit 1
fi
have_go="$(go env GOVERSION)"
have_go="${have_go#go}"
low="$(printf '%s\n' "$want_go" "$have_go" | sort -V | head -n1)"
if [[ "$low" != "$want_go" ]]; then
	echo "Go on PATH reports $have_go but go.mod requires $want_go or newer." >&2
	echo "Install Go $want_go+ (CI uses actions/setup-go with go-version-file: go.mod)." >&2
	exit 1
fi
echo "$(go version) (go.mod toolchain: $want_go)"

echo "== 2/8 gofmt =="
out="$(gofmt -l .)"
if [[ -n "$out" ]]; then
	echo "Unformatted files — run: gofmt -w on the paths below" >&2
	echo "$out" >&2
	exit 1
fi

echo "== 3/8 go mod verify =="
go mod verify

echo "== 4/8 go mod tidy (working tree must stay clean) =="
go mod tidy
git diff --exit-code go.mod go.sum

echo "== 5/8 go vet =="
go vet ./...

echo "== 6/8 golangci-lint (CI pins $GOLANGCI_PIN via scripts/golangci-version) =="
if ! command -v golangci-lint >/dev/null 2>&1; then
	echo "golangci-lint not on PATH. Install the same version CI uses:" >&2
	echo "  go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_PIN}" >&2
	echo "Docs: https://golangci-lint.run/welcome/install/" >&2
	exit 1
fi
if ! golangci-lint run ./... --timeout=5m; then
	echo "If the error mentions golangci-lint being built with an older Go than this module, reinstall with:" >&2
	echo "  go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_PIN}" >&2
	exit 1
fi

echo "== 7/8 go build ./... =="
go build ./...

echo "== 8/8 tests: unit (-short) then stdio smoke =="
go test -count=1 -race -short ./...
go test -count=1 -race ./cmd/service-manager-read-mcp -run '^TestCLI_stdioSmoke$'

echo "All checks passed."
