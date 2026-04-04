# Architecture

## Overview

`service-manager-read-mcp` is a stdio MCP server that answers read-only questions about service state by executing **allowlisted** platform CLIs and normalizing their output into shared domain types. Tool handlers live in `internal/servicemcp` and call `internal/detect` plus `internal/backend/*` only—no ad-hoc subprocesses.

## Threat model (short)

- **In scope:** constrain what an **MCP client** can trigger—only registered `CommandID` invocations with validated parameters, no shell, bounded output and time—so agents get **predictable read-only** supervisor queries instead of arbitrary `exec`.
- **Out of scope:** this is **not** a host sandbox. Anyone who can run the binary locally can already run `systemctl` (or equivalent); the design assumes you trust the environment where you install and invoke the server.
- **Residual risk:** backends call real supervisor tooling; supply-chain and parser bugs are mitigated by tests and allowlists but not formally verified here.

## Command execution boundaries

- **No shell:** the low-level `runner.Exec` backend uses `os/exec` with a resolved binary path and an explicit argv slice only.
- **No ad-hoc argv from tools:** higher layers call `runner.SafeRunner.RunInvocation` with a `runner.Invocation` that references a `runner.CommandID` registered in a `runner.Registry`.
- **Allowlist:** each `runner.Policy` pins the binary (resolved at registry construction), literal `FixedArgs`, and zero or more `ParamRule` slots. User- or agent-supplied data may only flow through validated parameter slots, never as flags.
- **Timeouts:** `runner.Spec.Timeout` is derived from policy defaults and optional per-invocation overrides. `exec.CommandContext` enforces cancellation.
- **Captured output:** stdout and stderr are streamed with per-stream byte caps (`DefaultMaxStreamBytes` when unset). Exceeding a cap aborts the process and returns `svcerr.ErrOutputLimit`.
- **Typed errors:** policy violations use `svcerr.ErrNotAllowlisted` and `svcerr.ErrInvalidArgument`. Process failures use `*svcerr.InvocationError` (unwraps to `svcerr.ErrInvocation`). Parser failures (backends) should use `svcerr.ParseError`.

## Backend interface contract

`internal/backend.ServiceManagerBackend` documents the cross-backend rules in code (context completion, `UnitRef.Manager` handling, typed errors, honest `Capabilities()`, normalized `domain` responses). Implementations return `ctx.Err()` when the context is already canceled or expired **before** other work—including before returning `svcerr.ErrUnsupported` for operations a backend does not implement.

`internal/detect.capabilityTemplate` is kept aligned with the concrete backends under `internal/backend/*`; `internal/detect/capabilities_test.go` asserts that alignment for “present on PATH” templates.

## Layering

| Layer | Responsibility |
| ----- | ---------------- |
| `internal/detect` | `GOOS`-gated `exec.LookPath` discovery for known supervisor CLIs; emits `Report` with `Finding` + `BackendCapabilities` templates. No subprocess execution. |
| `runner.Registry` / `SafeRunner` | Map `Invocation` → validated `Spec`; sole supported entry for untrusted inputs. |
| `runner.Runner` (`Exec`, `Func`) | Execute `Spec`; `Func` is the test double. |
| `internal/backend/systemd` | Allowlisted `systemctl` / `journalctl` recipes; parsers map JSON / key=value / journal short-iso into domain types. |
| `internal/backend/launchd` | Allowlisted `launchctl` (`version`, `list`, `print`); text/print parsers; failed list is list-based heuristic; logs unsupported. |
| `internal/backend/openrc` | Allowlisted `rc-service` (`-V`, `-l`, `<name> status`) and `rc-status --crashed`; text parsers; list has no runtime state; logs unsupported. |
| `internal/backend/runit` | Allowlisted `sv` (`-v`, `status <name>`); service-dir scan for list; logs unsupported. |
| `internal/backend/brew` | Allowlisted `brew` for `services list/info --json`; logs unsupported. |
| `internal/backend` (other) | Any future supervisors map domain operations to `CommandID` + params. |
| `internal/servicemcp` | MCP tool schemas and handlers (`list_service_managers`, `list_units`, `list_failed_services`, `get_service_status`, `get_recent_service_logs`); uses [github.com/modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk) with `StdioTransport`. Failed rows are sorted for stable output; see [docs/failed-services.md](failed-services.md). Log reads are line-bounded with MCP-side validation; see [docs/recent-logs.md](recent-logs.md). |
| `internal/version` | Release metadata (`Version`, `Commit`, `BuildDate`) and `MCPVersion()` for MCP `implementation.version`; see [docs/distribution.md](distribution.md). |
| `cmd/service-manager-read-mcp` | Process entry: `RunStdio` (logs to stderr; JSON-RPC on stdout); `-version` uses `internal/version.Summary`. |

## Testing strategy

- **Registry:** argv construction, unknown IDs, arity, and parameter validation (table-driven).
- **SafeRunner:** delegates to `Func` to assert the resolved `Spec` and to assert context deadlines propagate.
- **Exec:** uses tiny POSIX utilities when present (`true`, `false`, `echo`, `yes`, `sleep`) plus `readLimited` unit tests for capture logic.
- **MCP:** `internal/servicemcp` tests use in-memory transports for tool listing, schema presence, `tools/call`, and structured JSON outputs; `cmd/service-manager-read-mcp` has a short stdio smoke test (EOF stdin); `internal/e2e` drives `cmd/service-manager-read-mcp-e2e` over real stdio with a fake backend.

## Distribution

This repository is distributed as a **single Go binary** (no runtime beyond supervisor CLIs on `PATH`). Install, stdio MCP expectations, and release metadata are documented in [distribution.md](distribution.md).

## Related decisions

Recorded in [adr/README.md](adr/README.md) (read-only CLI model, stdio MCP, and ADR process).

## Releases and changelog

Maintainers: [releasing.md](releasing.md). User-visible history: [CHANGELOG.md](CHANGELOG.md).
