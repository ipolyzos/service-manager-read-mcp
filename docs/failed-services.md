# Failed services semantics (`list_failed_services`)

The MCP tool `list_failed_services` returns `domain.FailedServiceEntry` rows (`ref`, optional `reason`, optional `since`). All backends use the same JSON shape; **what counts as “failed” is not identical** across supervisors.

| Manager | Source | Meaning (summary) |
| ------- | ------ | ------------------- |
| **systemd** | `systemctl list-units --state=failed --type=service` (JSON) | Units systemd reports in **failed** state. Closest to a canonical “failed unit” list for this project. |
| **launchd** | `launchctl list` text columns | **Heuristic:** jobs whose **last exit status** column is non-zero. Not the same as a loaded failed launchd state; PID column and context still apply. |
| **OpenRC** | `rc-status --crashed` | Services OpenRC reports as **crashed** (requires a sufficiently new OpenRC; older hosts may error on invocation). |
| **runit** | Service scan + `sv status` | Services that are **wanted up** (`want` / `wants`) but **`sv status` reports down**. Different from systemd “failed” (no shared crash semantics). |
| **Homebrew** | `brew services list --json` | Formulas the parser treats as **error** / **other** non-healthy states from list JSON (see `internal/backend/brew`). |

## MCP behavior

- The tool requires `supports_list_failed_services` in the discovery snapshot for the chosen `manager`, same as other capability-gated tools.
- Results are sorted by `(ref.manager, ref.id, ref.display_name)` so repeated calls with the same host state produce **stable ordering**.
- An **empty** `failed` array means the backend reported no matching entries, not that the supervisor has no services.

## Related types

- [`domain.FailedServiceEntry`](../internal/domain/types.go)
- Backend implementations under `internal/backend/*/backend.go` (`ListFailedServices`)
