# service-manager-read-mcp

A **stdio** [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) server that exposes **read-only**, **structured** access to service state on Linux and macOS.

The server is designed to be **safe by default**: it does not start, stop, or reconfigure services. It wraps common service-manager command-line interfaces so agents and tools can query status and metadata in a consistent shape across platforms.

## MCP server (stdio)

The binary **`service-manager-read-mcp`** runs on **stdin/stdout** via the [Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk). Protocol JSON-RPC uses **stdout**; logs use **stderr**.

Tools return structured JSON (`structuredContent`). Call **`list_service_managers`** first to see which supervisors are on `PATH` and which `supports_*` capabilities apply.

| Tool | Input | Output |
| ---- | ----- | ------ |
| `list_service_managers` | _(none)_ | `report` — same shape as [detect.Report](internal/detect/detect.go) (`goos`, `findings[]` with `capabilities`) |
| `list_units` | `manager` (required): `systemd` \| `launchd` \| `openrc` \| `runit` \| `brew_services` (or `brew`); `limit` (optional, 0–10000) | `units[]` — [domain.Unit](internal/domain/types.go) rows |
| `list_failed_services` | `manager` (required), same values as `list_units` | `failed[]` — [domain.FailedServiceEntry](internal/domain/types.go); sorted for stable ordering; **meaning of “failed” varies by supervisor** — see [docs/failed-services.md](docs/failed-services.md) |
| `get_recent_service_logs` | `manager`, `service`, `max_lines` (0=default 100, max 500), optional `recent_window` (Go duration; empty/`0` = line tail only) | `logs[]` — [domain.RecentLogEntry](internal/domain/types.go); `effective_max_lines`, optional `truncated`. **Only systemd** has logs today; **non-zero `recent_window` is unsupported** until a safe time filter exists — see [docs/recent-logs.md](docs/recent-logs.md) |

`list_units`, `list_failed_services`, and `get_recent_service_logs` refuse managers that are not **present on `PATH`** for the current OS (see detection table below). Capabilities (`supports_list_failed_services`, `supports_get_recent_logs`, …) come from the same discovery report. All calls go through existing `internal/backend/*` code and the allowlisted runner.

`list_failed_services` arguments example:

```json
{ "manager": "systemd", "limit": 200 }
```

If discovery reports `supports_list_failed_services: false` for that manager, the tool returns a clear **unsupported** error (read-only; no shell).

`get_recent_service_logs` example (line-bounded tail only; no streaming/follow):

```json
{
  "manager": "systemd",
  "service": "nginx.service",
  "max_lines": 50,
  "recent_window": ""
}
```

### Build and run

Requires **[Go 1.25+](https://go.dev/dl/)** (MCP SDK).

```bash
go build -o bin/service-manager-read-mcp ./cmd/service-manager-read-mcp
./bin/service-manager-read-mcp
```

### IDE configuration (example)

Point your MCP client at the built binary (absolute path). Example shape:

```json
{
  "mcpServers": {
    "service-manager-read": {
      "command": "/absolute/path/to/bin/service-manager-read-mcp",
      "args": []
    }
  }
}
```

Adjust keys to match your client (Cursor, Claude Desktop, etc.).

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

Each finding includes a `domain.BackendCapabilities` template describing **planned** read-only surface area for that CLI; for managers with a library backend in this repo, the template is kept consistent with `internal/backend/<name>.Capabilities()`.

Shared rules for all backends (context cancellation, `UnitRef.Manager`, typed errors) live on [`ServiceManagerBackend`](internal/backend/backend.go). See [docs/architecture.md](docs/architecture.md) § *Backend interface contract*.

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

## Launchd backend

Package `internal/backend/launchd` implements `ServiceManagerBackend` via **`launchd.New()`** and the same allowlisted runner. Only `launchctl` is used (`version`, `list`, `print <target>`).

| Operation | CLI (summarized) |
| --------- | ---------------- |
| `DetectServiceManager` | `launchctl version` (requires **macOS** / `GOOS=darwin`; otherwise `ErrBackendUnavailable`) |
| `ListUnits` | `launchctl list` (text columns) |
| `GetServiceStatus` | `launchctl print` on `gui/<uid>/`, then `user/<uid>/`, then `system/` targets |
| `ListFailedServices` | Same `list` output; **heuristic:** non-zero **last exit status** column |
| `GetRecentServiceLogs` | **`ErrUnsupported`** — no safe allowlisted equivalent to per-unit `journalctl`; `Capabilities().SupportsRecentLogs` is **false** |

### Semantic gaps vs systemd

| Topic | systemd backend | launchd backend |
| ----- | ---------------- | ---------------- |
| Unit identity | Unit names (e.g. `foo.service`) | Job **labels** (`com.example.app`) |
| Listing | Typed JSON, all listed service units | Plain `list` for the invoking context; coarse **running** inferred from PID column only |
| Failed services | Explicit `failed` state from systemd | **Approximation** from list’s last exit status ≠ 0 (not the same as “loaded failed unit”) |
| Detail view | Rich `systemctl show` properties | Best-effort scan of `launchctl print` key/value lines |
| Logs | Bounded `journalctl` tail | Not exposed |

Correctness prefers **honest mapping** and documented limits over forcing identical semantics.

## OpenRC backend

Package `internal/backend/openrc` implements `ServiceManagerBackend` via **`openrc.New()`** using allowlisted **`rc-service`** and **`rc-status`** only.

| Operation | CLI (summarized) |
| --------- | ---------------- |
| `DetectServiceManager` | `rc-service -V` (expects “OpenRC” in stdout) |
| `ListUnits` | `rc-service -l` (one name per line) |
| `GetServiceStatus` | `rc-service <name> status` (stdout+stderr scanned for started/stopped/crashed) |
| `ListFailedServices` | `rc-status --crashed` (OpenRC **≥ ~0.45**; older releases return an invocation error) |
| `GetRecentServiceLogs` | **`ErrUnsupported`** — no safe per-service log tail; `SupportsRecentLogs` is **false** |

### Differences vs systemd

| Topic | systemd | OpenRC (`rc-service`) |
| ----- | -------- | --------------------- |
| Listing | JSON with load/active/sub | **Names only**; `Unit.State` is **`unknown`** until `GetServiceStatus` |
| Status | Structured `systemctl show` | Heuristic string match on `status` output; non-zero exit may still mean “stopped” (handled) |
| Failed | `systemctl` failed units | **`rc-status --crashed`** only — not the same as systemd’s failure model |
| Logs | `journalctl` | Not implemented here |

## Project layout

```
cmd/service-manager-read-mcp/   # MCP stdio entrypoint (+ integration smoke test)
internal/
  servicemcp/                   # MCP tool schemas, handlers, RegisterTools, RunStdio
  domain/                       # shared types (units, status, logs, capabilities)
  svcerr/                       # typed errors for backends and runners
  runner/                       # argv-only subprocess runner (+ test fake)
  detect/                       # PATH-based supervisor discovery (no subprocess)
  backend/                      # ServiceManagerBackend interface
  backend/stub/                 # noop backend for wiring tests
  backend/systemd/              # systemd/journald backend (allowlisted CLIs)
  backend/launchd/              # launchd backend (allowlisted launchctl)
  backend/openrc/               # OpenRC backend (rc-service / rc-status)
  backend/runit/                # runit backend (allowlisted sv)
  backend/brew/                 # Homebrew services backend (allowlisted brew)
```

## Development

Requires [Go 1.25+](https://go.dev/dl/) for the MCP SDK.

```bash
go build -o bin/service-manager-read-mcp ./cmd/service-manager-read-mcp
go test ./...
go vet ./...
```

Subprocess execution for backends is routed through `internal/runner`: build an allowlisted `runner.Registry`, wrap a `runner.Exec` (or test `runner.Func`) with `runner.SafeRunner`, and call `RunInvocation`—never hand-construct argv from tool or agent input. MCP tools call `internal/servicemcp` only, which uses `detect` and `backend` packages. See [docs/architecture.md](docs/architecture.md).

## Status

MCP **stdio server** with tools **`list_service_managers`** and **`list_units`** is wired. Library backends (**systemd**, **launchd**, **OpenRC**, **runit**, **Homebrew services**) and allowlisted execution remain the only way to reach host CLIs. Additional MCP tools and supervisors may follow.

## License

Apache License 2.0. See [LICENSE](LICENSE) and [NOTICE](NOTICE).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). The project uses a [Code of Conduct](CODE_OF_CONDUCT.md).
