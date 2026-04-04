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

## Systemd backend

Package `internal/backend/systemd` implements `ServiceManagerBackend` using **`systemd.New()`** and the shared allowlisted runner (`SafeRunner` + `Exec`). All argv is defined in code; unit names and journal line counts are validated before invocation.

| Operation | CLI (summarized) |
| --------- | ---------------- |
| `DetectServiceManager` | `systemctl show systemd --property=Id` |
| `ListUnits` | `systemctl list-units --type=service --all -o json` |
| `GetServiceStatus` | `systemctl show <unit>` (fixed property set) |
| `ListFailedServices` | `systemctl list-units --state=failed --type=service -o json` |
| `GetRecentServiceLogs` | `journalctl -o short-iso -n <N> -u <unit>` |

**Limits:** log tail uses `MaxLines` (default 100, capped at 500). **`RecentLogsInput.Since` is not implemented** and returns `ErrUnsupported` until a safe allowlisted `--since` path exists.

**Parsing:** list/failed output expects `systemctl` JSON (array or `{ "units": [...] }`). Journal lines use `-o short-iso`; severities are left as `unknown`. Hosts need compatible systemd/journald versions; failures surface as invocation or parse errors.

## Runit backend

Package `internal/backend/runit` implements `ServiceManagerBackend` via **`runit.New()`** using allowlisted **`sv`** plus a **read-only directory listing** of the runit service tree (names are immediate children of the scan directory).

| Operation | Mechanism (summarized) |
| --------- | ------------------------ |
| `DetectServiceManager` | `sv -v` stdout must look like **runit** (not BusyBox); rejects ambiguous `sv` implementations |
| `ListUnits` | `os.ReadDir` on **`SVDIR`** if set and valid, else first of `/var/service`, `/service`, `/etc/service`; `Unit.State` is **`unknown`** until `GetServiceStatus` |
| `GetServiceStatus` | `sv status <name>` with subprocess **`Dir`** set to the resolved service directory |
| `ListFailedServices` | One `sv status` per listed service; **heuristic:** line matching `down: <name>: … want up` (supervise wants the service up but it is down) |
| `GetRecentServiceLogs` | **`ErrUnsupported`** — no bounded, allowlisted per-service log tail via `sv`; `SupportsRecentLogs` is **false** |

### Semantic gaps vs systemd

| Topic | systemd | runit (`sv`) |
| ----- | -------- | ------------- |
| Unit list source | `systemctl` JSON | **Filesystem directory** + name validation; not the same as “all units systemd knows” |
| List row state | Load/active/sub states | **`unknown`** until a status call |
| Status | `systemctl show` properties | Heuristic parse of **`sv status`** text; PID from `(pid N)` when present |
| Failed services | Explicit failed units | **`down … want up` only** — ignores clean “down, normally up” and does not see other failure classes |
| Logs | Bounded `journalctl` | Not exposed here |

`ListFailedServices` may be **slow** on large service directories because it runs **`sv status` once per service**.

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
  backend/systemd/              # systemd/journald backend (allowlisted CLIs)
  backend/runit/                # runit backend (allowlisted sv + service dir scan)
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

Domain model, backend interface, allowlisted subprocess runner, PATH-based supervisor detection, a stub backend, **systemd**, and **runit** backends (library only) are in place. MCP stdio wiring and other supervisor backends are not implemented yet.

## License

Apache License 2.0. See [LICENSE](LICENSE) and [NOTICE](NOTICE).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). The project uses a [Code of Conduct](CODE_OF_CONDUCT.md).
