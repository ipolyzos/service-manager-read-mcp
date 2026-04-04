# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

### Changed

### Fixed

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
- `SECURITY.md`, `CHANGELOG.md`, `docs/releasing.md`, `docs/distribution.md`, `internal/version`, and architecture decision records (`docs/adr/`).

[Unreleased]: https://github.com/ipolyzos/service-manager-read-mcp/compare/v0.1.1...HEAD
[0.1.1]: https://github.com/ipolyzos/service-manager-read-mcp/releases/tag/v0.1.1
[0.1.0]: https://github.com/ipolyzos/service-manager-read-mcp/releases/tag/v0.1.0
