# Notes for agents and maintainers

## Branch flow

Open pull requests against **`develop`**. **`main`** tracks stable releases.

## CI parity

GitHub Actions (`.github/workflows/ci.yml`) mirrors **`make ci`** / **`./scripts/ci.sh`**:

| Step | Command / action |
| ---- | ---------------- |
| Format | `gofmt -l .` must be empty |
| Modules | `go mod verify`; `go mod tidy` then no diff on `go.mod` / `go.sum` |
| Vet | `go vet ./...` |
| Lint | `golangci-lint run ./...` (see `.golangci.yml`) |
| Build | `go build ./...` |
| Unit | `go test -race -short ./...` (skips stdio smoke in `cmd/service-manager-read-mcp`) |
| Stdio smoke | `go test -race ./cmd/service-manager-read-mcp -run '^TestCLI_stdioSmoke$'` |

**Build**, **unit**, and **stdio smoke** run on **ubuntu-latest** and **macos-latest** after **static** succeeds (`needs: static`), so formatter or lint failures skip matrix work.

The **static** job runs only on Ubuntu to keep signal fast and avoid duplicating lint downloads.
