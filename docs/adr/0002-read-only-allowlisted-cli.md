# ADR 0002: Read-only host access via allowlisted argv

## Status

Accepted

## Context

Agents need structured service state from real hosts. Running arbitrary shell or unconstrained argv from tool input is unsafe.

## Decision

- **Read-only:** MCP tools never start, stop, or reconfigure services.
- **No shell:** subprocesses use `os/exec` with explicit argv only.
- **Allowlist:** `internal/runner` maps registered `CommandID` values to fixed binaries, literal args, and validated parameter slots. User-supplied identifiers pass backend-specific validation before any argv is built.

## Consequences

- New supervisor operations require registry entries and parsers, not ad-hoc commands.
- Some platforms (e.g. launchd logs) remain unsupported until a safe mapping exists.
