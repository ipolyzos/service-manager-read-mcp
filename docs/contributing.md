# Contributing

Thank you for your interest in **service-manager-read-mcp**.

## Workflow

- Open issues for bugs and feature discussion before large changes when practical.
- Use short-lived branches with one of these prefixes: `feature/`, `bugfix/`, or `chore/`.
- **Do not push commits directly to `develop` or `main`.** Push your topic branch and open a **GitHub Pull Request** into **`develop`** so required CI runs and review can happen.
- Keep changes small, focused, and easy to review.
- Update user-facing documentation when behavior, configuration, or packaging changes (see **[docs/README.md](README.md)** for the doc map).

## Releases

`main` is used for stable releases. Work merges into **`develop`** via PR; when cutting a release, promote **`develop` → `main`** with a **GitHub Pull Request** as well (then tag on **`main`** after that merge). See **[releasing.md](releasing.md)**.

Steps, tagging, and `go install`: **[releasing.md](releasing.md)**. Update **[CHANGELOG.md](CHANGELOG.md)** on your PR branch before the release PR merges. Prefer a green **`make ci`** locally when preparing a release.

## Distribution

End-user install options, stdio MCP usage, `go install`, and **`make build-release`** are documented in **[distribution.md](distribution.md)**.

## Security

Report vulnerabilities per **[security.md](security.md)** (not public issues for undisclosed problems).

## Conduct

All participants are expected to follow the [Code of Conduct](code-of-conduct.md).

## Testing

From the repository root (requires Go **1.25+**, same as [go.mod](../go.mod)):

```bash
go test ./...
```

Stdio MCP end-to-end coverage (builds `cmd/service-manager-read-mcp-e2e`, then drives it over stdin/stdout with the official MCP client transport):

```bash
go test ./internal/e2e/...
```

Quick test sweep (CI also runs `gofmt`, modules, `vet`, **golangci-lint**, build, `-short` tests, and the stdio smoke test—see **`make ci`**):

```bash
go test -race ./...
```

Full static + build + test parity (requires **golangci-lint** on `PATH`): **`make ci`** or **`./scripts/ci.sh`** — see **[agents.md](agents.md)**.

To exercise the **real** installed MCP binary against **systemd/journalctl** on your machine (all five tools plus journal edge cases), after `go install` or `make build`:

```bash
go run ./scripts/host_mcp_smoke.go -q
```

Default binary resolution: **`SERVICE_MANAGER_READ_MCP_BIN`** if set; else **`service-manager-read-mcp` on `PATH`**; else **`$HOME/go/bin/service-manager-read-mcp`**. Override with a path argument: `go run ./scripts/host_mcp_smoke.go -q /path/to/service-manager-read-mcp`. Omit `-q` for full JSON per tool.

## License

By contributing, you agree that your contributions are licensed under the Apache License 2.0, the same license as the project. See [LICENSE](../LICENSE) and [NOTICE](../NOTICE).
