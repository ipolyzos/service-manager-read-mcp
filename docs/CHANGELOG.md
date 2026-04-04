# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

### Changed

### Fixed

## [0.1.5] — 2026-04-04

### Added

- **Threat model** blurb on the root **[README.md](../README.md)** and a short **Threat model** section in **[architecture.md](architecture.md)** (read-only scope, no host sandbox, residual risks).

### Changed

- **GitHub Actions** **`.github/workflows/ci.yml`** now matches **`./scripts/ci.sh`** / **`make ci`**: `gofmt`, `go mod verify`, `go mod tidy` with a clean `go.mod`/`go.sum` diff, `go vet`, **golangci-lint** (pinned **v2.11.4**), `go build ./...`, `go test -race -short ./...`, and the stdio smoke test on **`cmd/service-manager-read-mcp`**.
- **[agents.md](agents.md)**, **[mcp-tools.md](mcp-tools.md)**, **[contributing.md](contributing.md)**, **`scripts/ci.sh`**, and the **Makefile** describe the same CI sequence end-to-end.

## [0.1.4] — 2026-04-04

### Changed

- Move long-form README content into **docs/** ([mcp-tools.md](mcp-tools.md), [supervisors.md](supervisors.md), [repository-layout.md](repository-layout.md)); root **README.md** is pointers only. Refresh [docs/README.md](README.md) hub.

## [0.1.3] — 2026-04-04

### Changed

- Document **GitHub Pull Request**–only integration: topic branch → **`develop`**, then **`develop` → `main`** for releases (see [contributing.md](contributing.md), [releasing.md](releasing.md), [agents.md](agents.md)).

## [0.1.2] — 2026-04-04

### Added

- **`scripts/host_mcp_smoke.go`**: optional local driver (stdio MCP) to smoke-test an installed `service-manager-read-mcp` binary against the real host; documented in **[contributing.md](contributing.md)**.
- **`docs/README.md`**: documentation hub linking install, architecture, ADRs, and community docs.

### Changed

- Consolidated contributor-facing markdown under **`docs/`** (**contributing**, **security**, **code of conduct**, **changelog**); root files remain short stubs for GitHub discovery and stable URLs.

### Fixed

- **systemd logs:** tolerate `journalctl -o short-iso` output when the timestamp field is `--` (unknown time) and skip the `-- No entries --` sentinel so `get_recent_service_logs` does not fail on empty tails or mixed journal lines.
- **`host_mcp_smoke`:** default binary resolves via **`SERVICE_MANAGER_READ_MCP_BIN`**, then **`service-manager-read-mcp` on `PATH`**, then **`$HOME/go/bin/service-manager-read-mcp`** (still overridable with a path argument).

## [0.1.1] — 2026-04-04

### Fixed

- Remove stray Git merge conflict markers that were accidentally left in `README.md` (published tree for v0.1.0).

## [0.1.0] — 2026-04-04

First public release: stdio MCP server for **read-only** service supervisor state on Linux and macOS.

### Added

- MCP tools: `list_service_managers`, `list_units`, `list_failed_services`, `get_service_status`, `get_recent_service_logs` (capability-gated per supervisor).
- Backends: systemd, launchd, OpenRC, runit, Homebrew services (`internal/backend/*`) via allowlisted `internal/runner`.
- PATH-based discovery (`internal/detect`) without invoking supervisor CLIs during detection.
- CI (build + race tests), `make ci` / `scripts/ci.sh`, and stdio subprocess smoke plus `internal/e2e` MCP client tests.
- Security policy, changelog, `docs/releasing.md`, `docs/distribution.md`, `internal/version`, and architecture decision records (`docs/adr/`).

[Unreleased]: https://github.com/ipolyzos/service-manager-read-mcp/compare/v0.1.5...HEAD
[0.1.5]: https://github.com/ipolyzos/service-manager-read-mcp/releases/tag/v0.1.5
[0.1.4]: https://github.com/ipolyzos/service-manager-read-mcp/releases/tag/v0.1.4
[0.1.3]: https://github.com/ipolyzos/service-manager-read-mcp/releases/tag/v0.1.3
[0.1.2]: https://github.com/ipolyzos/service-manager-read-mcp/releases/tag/v0.1.2
[0.1.1]: https://github.com/ipolyzos/service-manager-read-mcp/releases/tag/v0.1.1
[0.1.0]: https://github.com/ipolyzos/service-manager-read-mcp/releases/tag/v0.1.0
