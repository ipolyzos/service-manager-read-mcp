# MCP tools (stdio server)

The binary **`service-manager-read-mcp`** runs on **stdin/stdout** via the [Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk). **JSON-RPC uses stdout only**; **stderr** is for server logs. Do not wrap stdout in a PTY or mix other streams into it.

**Client spawn rule:** use **`args`: `[]`** (or omit `args`). **`-version`** is only for humans and exits without starting MCP.

Tools return structured JSON (`structuredContent`). Call **`list_service_managers`** first to see which supervisors are on `PATH` and which `supports_*` capabilities apply.

| Tool | Input | Output |
| ---- | ----- | ------ |
| `list_service_managers` | _(none)_ | `report` — same shape as [detect.Report](../internal/detect/detect.go) (`goos`, `findings[]` with `capabilities`) |
| `list_units` | `manager` (required): `systemd` \| `launchd` \| `openrc` \| `runit` \| `brew_services` (or `brew`); `limit` (optional, 0–10000) | `units[]` — [domain.Unit](../internal/domain/types.go) rows |
| `list_failed_services` | `manager` (required), same values as `list_units` | `failed[]` — [domain.FailedServiceEntry](../internal/domain/types.go); sorted for stable ordering; **meaning of “failed” varies by supervisor** — see [failed-services.md](failed-services.md) |
| `get_service_status` | `manager` (required), `id` (unit/service id for that supervisor) | `status` — [domain.ServiceStatus](../internal/domain/types.go) |
| `get_recent_service_logs` | `manager`, `service`, `max_lines` (0=default 100, max 500), optional `recent_window` (Go duration; empty/`0` = line tail only) | `logs[]` — [domain.RecentLogEntry](../internal/domain/types.go); `effective_max_lines`, optional `truncated`. **Only systemd** has logs today; **non-zero `recent_window` is unsupported** until a safe time filter exists — see [recent-logs.md](recent-logs.md) |

`list_units`, `list_failed_services`, `get_service_status`, and `get_recent_service_logs` refuse managers that are not **present on `PATH`** for the current OS (see [supervisors.md](supervisors.md) § *Service manager detection*). Capabilities (`supports_get_service_status`, `supports_list_failed_services`, `supports_get_recent_logs`, …) come from the same discovery report. All calls go through existing `internal/backend/*` code and the allowlisted runner.

`list_failed_services` arguments example:

```json
{ "manager": "systemd" }
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

## Build, install, CI (summary)

Requires **[Go 1.25+](https://go.dev/dl/)** (MCP SDK).

```bash
make build
./bin/service-manager-read-mcp -version
./bin/service-manager-read-mcp
```

After the first **semver tag**, install with:

```bash
go install github.com/ipolyzos/service-manager-read-mcp/cmd/service-manager-read-mcp@latest
```

(See [releasing.md](releasing.md) for tagging.)

**GitHub Actions** runs the same sequence as **`make ci`** / **`./scripts/ci.sh`** (see **[agents.md](agents.md)**). Locally you need **[golangci-lint](https://golangci-lint.run/welcome/install/)** on `PATH`. See **[contributing.md](contributing.md)** for focused targets (`make test-short`, `make test-smoke`).

## MCP client configuration

Use an **absolute path** and empty **`args`**. Full examples and version metadata: **[distribution.md](distribution.md)** § *MCP client configuration*.

## See also

- [supervisors.md](supervisors.md) — planned scope, PATH detection, per-backend CLI details  
- [architecture.md](architecture.md) — layering and safety model  
- [README.md](../README.md) — repository entrypoint (pointers only)
