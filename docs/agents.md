# Notes for agents and maintainers

## Branch flow

- Land work with a **GitHub Pull Request** into **`develop`** from `feature/*`, `bugfix/*`, or `chore/*` (no direct pushes to **`develop`** / **`main`**).
- Promote releases with a **Pull Request** from **`develop`** into **`main`**, then tag on **`main`** (see [releasing.md](releasing.md)).

## CI vs local checks

**GitHub Actions** (`.github/workflows/ci.yml`) on push/PR to `main` and `develop`:

| Step | Command |
| ---- | ------- |
| Build | `go build ./...` |
| Test | `go test -race ./...` |

**Local `make ci`** / **`./scripts/ci.sh`** is stricter: `gofmt`, `go mod verify`, `go mod tidy` check, `go vet`, `golangci-lint`, `go build ./...`, `go test -race -short ./...`, then the stdio smoke test. Run it before release prep or when changing Go code.

## Releases

See [releasing.md](releasing.md) and [CHANGELOG.md](CHANGELOG.md).
