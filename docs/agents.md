# Notes for agents and maintainers

## Branch flow

- Land work with a **GitHub Pull Request** into **`develop`** from `feature/*`, `bugfix/*`, or `chore/*` (no direct pushes to **`develop`** / **`main`**).
- Promote releases with a **Pull Request** from **`develop`** into **`main`**, then tag on **`main`** (see [releasing.md](releasing.md)).

## CI vs local checks

**GitHub Actions** runs **`.github/workflows/ci-verify.yml`** (reusable). **`.github/workflows/ci.yml`** calls it on push/PR to **`main`** and **`develop`**; **`.github/workflows/release.yml`** calls it on **`v*`** tags before creating a GitHub Release. The sequence matches **`./scripts/ci.sh`** / **`make ci`**:

**golangci-lint** version is pinned in **`scripts/golangci-version`** (CI reads that file; see **[contributing.md](contributing.md)** — Local toolchain).

| Step | Command |
| ---- | ------- |
| Format | `gofmt -l` (must be empty) |
| Modules | `go mod verify`; `go mod tidy` then `git diff --exit-code go.mod go.sum` |
| Vet | `go vet ./...` |
| Lint | `golangci-lint run ./... --timeout=5m` (version from `scripts/golangci-version`) |
| Build | `go build ./...` |
| Test | `go test -count=1 -race -short ./...` |
| Smoke | `go test -count=1 -race ./cmd/service-manager-read-mcp -run '^TestCLI_stdioSmoke$'` |

Run **`make ci`** locally before release prep or when changing Go code (requires **golangci-lint** on **`PATH`**, same version as **`scripts/golangci-version`**).

## Releases

See [releasing.md](releasing.md) and [CHANGELOG.md](CHANGELOG.md).
