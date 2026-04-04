# Supervisors: scope, detection, and backends

## Scope (planned)

Platform tooling this project intends to wrap (where available on the host):

| Platform / ecosystem | Typical CLI |
| -------------------- | ----------- |
| systemd (Linux) | `systemctl` |
| launchd (macOS) | `launchctl` |
| OpenRC | `rc-service` |
| runit | `sv` |
| Homebrew (Linux/macOS) | `brew` / `brew services` |

## Service manager detection

`internal/detect` reports which supervisor CLIs are **visible on `PATH`** for the current host (`GOOS`-gated). It uses **executable resolution only** (via `exec.LookPath`); it does **not** invoke `systemctl`, `launchctl`, `brew`, or other tools during detection.

| Manager | Binary looked up | Applicable OS targets (coarse) |
| ------- | ---------------- | ------------------------------- |
| systemd | `systemctl` | Linux |
| launchd | `launchctl` | macOS |
| OpenRC | `rc-service` | Linux, FreeBSD |
| runit | `sv` | Linux, BSDs, macOS (if installed) |
| Homebrew | `brew` | Linux, macOS (Homebrew installs) |

Each finding includes a `domain.BackendCapabilities` template describing **planned** read-only surface area for that CLI; for managers with a library backend in this repo, the template is kept consistent with `internal/backend/<name>.Capabilities()`.

Shared rules for all backends (context cancellation, `UnitRef.Manager`, typed errors) live on [`ServiceManagerBackend`](../internal/backend/backend.go). See [architecture.md](architecture.md) § *Backend interface contract*.

**Runit** and **Homebrew services** backends follow the same allowlisted-runner model; command recipes and parsers are summarized in [architecture.md](architecture.md) § *Layering*.

---

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

---

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

---

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
