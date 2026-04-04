# service-manager-read-mcp

A **stdio** [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) server that exposes **read-only**, **structured** access to service state on Linux and macOS.

The server is designed to be **safe by default**: it does not start, stop, or reconfigure services. It wraps common service-manager command-line interfaces so agents and tools can query status and metadata in a consistent shape across platforms.

## Scope (planned)

Platform tooling this project intends to wrap (where available on the host):

| Platform / ecosystem | Typical CLI |
| ---------------------- | ----------- |
| systemd (Linux)      | `systemctl` |
| launchd (macOS)      | `launchctl` |
| OpenRC               | `rc-service` |
| runit                | `sv`        |
| Homebrew (Linux/macOS) | `brew` / `brew services` |

## Service manager detection

`internal/detect` reports which supervisor CLIs are **visible on `PATH`** for the current host (`GOOS`-gated). It uses **executable resolution only** (via `exec.LookPath`); it does **not** invoke `systemctl`, `launchctl`, `brew`, or other tools during detection.

| Manager   | Binary looked up | Applicable OS targets (coarse)      |
| --------- | ---------------- | ------------------------------------ |
| systemd   | `systemctl`      | Linux                                |
| launchd   | `launchctl`      | macOS                                |
| OpenRC    | `rc-service`     | Linux, FreeBSD                       |
| runit     | `sv`             | Linux, BSDs, macOS (if installed)    |
| Homebrew  | `brew`           | Linux, macOS (Homebrew installs)     |

Each finding includes a `domain.BackendCapabilities` template describing **planned** read-only surface area once a backend exists; presence on PATH does not imply the MCP server can query it yet.

## Project layout

```
cmd/service-manager-read-mcp/   # process entrypoint (MCP wiring pending)
internal/
  domain/                       # shared types (units, status, logs, capabilities)
  svcerr/                       # typed errors for backends and runners
  runner/                       # argv-only subprocess runner (+ test fake)
  detect/                       # PATH-based supervisor discovery (no subprocess)
  backend/                      # ServiceManagerBackend interface
  backend/stub/                 # noop backend for wiring tests
```

## Development

Requires [Go 1.22+](https://go.dev/dl/).

```bash
go build -o bin/service-manager-read-mcp ./cmd/service-manager-read-mcp
go test ./...
go vet ./...
```

Subprocess execution for backends is routed through `internal/runner`: build an allowlisted `runner.Registry`, wrap a `runner.Exec` (or test `runner.Func`) with `runner.SafeRunner`, and call `RunInvocation`—never hand-construct argv from tool or agent input. See [docs/architecture.md](docs/architecture.md).

## Status

Domain model, backend interface, allowlisted subprocess runner, PATH-based supervisor detection, and a stub backend are in place. MCP stdio wiring and real supervisor integrations are not implemented yet.

## License

Apache License 2.0. See [LICENSE](LICENSE) and [NOTICE](NOTICE).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). The project uses a [Code of Conduct](CODE_OF_CONDUCT.md).
