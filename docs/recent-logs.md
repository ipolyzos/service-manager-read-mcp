# Recent service logs (`get_recent_service_logs`)

The MCP tool `get_recent_service_logs` returns a **bounded, non-streaming** list of log lines as [`domain.RecentLogEntry`](../internal/domain/types.go) (`timestamp` when parseable, `line`, `severity`, optional `unit`). There is **no follow mode**, **no streaming**, and **no unbounded reads** at the MCP layer.

## Request fields (all bounded)

| Field | Meaning |
| ----- | ------- |
| `manager` | Same supervisor key as other tools (`systemd`, `launchd`, …). |
| `service` | Unit / service identifier for that manager (same validation as backend argv). |
| `max_lines` | Maximum lines to return. **0** means default **100**. Hard cap **500** at the MCP boundary. |
| `recent_window` | Optional Go [`time.ParseDuration`](https://pkg.go.dev/time#ParseDuration) string (e.g. `1h`, `30m`). **Empty** or **`0`** means *line-bounded tail only* (no time window). Parsed durations are capped at **7 days** (`168h`). |

## Current backend support

| Manager | `supports_get_recent_logs` | Behavior |
| ------- | -------------------------- | -------- |
| **systemd** | yes | Allowlisted `journalctl` tail (`-n`, `-u`). `RecentLogsInput.Since` is **not** wired; **`recent_window` must be empty or `0`** — non-zero returns `ErrUnsupported` from the MCP handler until a safe `--since` path is allowlisted. |
| **launchd**, **OpenRC**, **runit**, **brew** | no | Tool fails with **`operation unsupported`** when discovery reports logs off (read-only; no shell). |

## Truncation

If a backend ever returns more lines than the effective `max_lines`, the handler keeps the **last** N lines (newest tail segment) and sets `truncated: true` in the result. Normal systemd paths should not require this.

## Related code

- MCP validation: `internal/servicemcp/logs_validate.go`, `ParseRecentLogsBounds`
- Backend: `internal/backend/systemd` (`GetRecentServiceLogs`, `ParseJournalShortISO`)
