# Architecture

## Overview

`service-manager-read-mcp` is a stdio MCP server (wiring pending) that answers read-only questions about service state by executing **allowlisted** platform CLIs and normalizing their output into shared domain types.

## Command execution boundaries

- **No shell:** the low-level `runner.Exec` backend uses `os/exec` with a resolved binary path and an explicit argv slice only.
- **No ad-hoc argv from tools:** higher layers call `runner.SafeRunner.RunInvocation` with a `runner.Invocation` that references a `runner.CommandID` registered in a `runner.Registry`.
- **Allowlist:** each `runner.Policy` pins the binary (resolved at registry construction), literal `FixedArgs`, and zero or more `ParamRule` slots. User- or agent-supplied data may only flow through validated parameter slots, never as flags.
- **Timeouts:** `runner.Spec.Timeout` is derived from policy defaults and optional per-invocation overrides. `exec.CommandContext` enforces cancellation.
- **Captured output:** stdout and stderr are streamed with per-stream byte caps (`DefaultMaxStreamBytes` when unset). Exceeding a cap aborts the process and returns `svcerr.ErrOutputLimit`.
- **Typed errors:** policy violations use `svcerr.ErrNotAllowlisted` and `svcerr.ErrInvalidArgument`. Process failures use `*svcerr.InvocationError` (unwraps to `svcerr.ErrInvocation`). Parser failures (backends) should use `svcerr.ParseError`.

## Layering

| Layer | Responsibility |
| ----- | ---------------- |
| `internal/detect` | `GOOS`-gated `exec.LookPath` discovery for known supervisor CLIs; emits `Report` with `Finding` + `BackendCapabilities` templates. No subprocess execution. |
| `runner.Registry` / `SafeRunner` | Map `Invocation` → validated `Spec`; sole supported entry for untrusted inputs. |
| `runner.Runner` (`Exec`, `Func`) | Execute `Spec`; `Func` is the test double. |
| `internal/backend/systemd` | Allowlisted `systemctl` / `journalctl` recipes; parsers map JSON / key=value / journal short-iso into domain types. |
| `internal/backend/launchd` | Allowlisted `launchctl` (`version`, `list`, `print`); text/print parsers; failed list is list-based heuristic; logs unsupported. |
| `internal/backend` (other) | Additional supervisors map domain operations to `CommandID` + params. |
| MCP surface (future) | JSON-RPC over stdio; maps tool calls to backend operations. |

## Testing strategy

- **Registry:** argv construction, unknown IDs, arity, and parameter validation (table-driven).
- **SafeRunner:** delegates to `Func` to assert the resolved `Spec` and to assert context deadlines propagate.
- **Exec:** uses tiny POSIX utilities when present (`true`, `false`, `echo`, `yes`, `sleep`) plus `readLimited` unit tests for capture logic.

## Related decisions

See `docs/adr/` for language choice, backend abstraction, and the read-only safety model.
