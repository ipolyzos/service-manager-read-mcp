# Recent service logs (`get_recent_service_logs`)

The MCP tool `get_recent_service_logs` returns a **bounded, non-streaming** list of log lines as [`domain.RecentLogEntry`](../internal/domain/types.go) (`timestamp` when parseable, `line`, `severity`, optional `unit`). There is **no follow mode**, **no streaming**, and **no unbounded reads** at the MCP layer.

## Request fields (all bounded)

| Field | Meaning |
| ----- | ------- |
| `manager` | Same supervisor key as other tools (`systemd`, `launchd`, …). |
| `service` | Unit / service identifier for that manager (same validation as backend argv). |
| `max_lines` | Maximum lines to return. **0** means default **100**. Hard cap **500** at the MCP boundary. |
| `recent_window` | **Use empty or `0` only.** That selects a **line-bounded tail** (`max_lines` / journal `-n`); no time filter is applied today. A non-empty, non-zero value is checked by [`ParseRecentLogsBounds`](../internal/servicemcp/logs_validate.go) (valid Go duration syntax; values above **7 days** / `168h` are rejected as invalid), but the MCP handler then returns **`unsupported`** — there is **no** allowlisted `journalctl --since` (or equivalent) path yet, so **no** time window actually runs. |

## Current backend support

| Manager | `supports_get_recent_logs` | Behavior |
| ------- | -------------------------- | -------- |
| **systemd** | yes | Allowlisted `journalctl` tail (`-n`, `-u`). Time-bounded windows are **not** implemented; see **`recent_window`** in the table above. |
| **launchd**, **OpenRC**, **runit**, **brew** | no | Tool fails with **`operation unsupported`** when discovery reports logs off (read-only; no shell). |

## Truncation

If a backend ever returns more lines than the effective `max_lines`, the handler keeps the **last** N lines (newest tail segment) and sets `truncated: true` in the result. Normal systemd paths should not require this.

## Related code

- MCP validation: `internal/servicemcp/logs_validate.go`, `ParseRecentLogsBounds`
- Backend: `internal/backend/systemd` (`GetRecentServiceLogs`, `ParseJournalShortISO`)
