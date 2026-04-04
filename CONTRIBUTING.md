# Contributing

Thank you for your interest in **service-manager-read-mcp**.

## Workflow

- Open issues for bugs and feature discussion before large changes when practical.
- Use short-lived branches with one of these prefixes: `feature/`, `bugfix/`, or `chore/`.
- Open pull requests against the **`develop`** branch (not `main`).
- Keep changes small, focused, and easy to review.
- Update user-facing documentation when behavior, configuration, or packaging changes.

## Releases

`main` is used for stable releases. Changes integrate on `develop` and are promoted to `main` when cutting a release.

## Conduct

All participants are expected to follow the [Code of Conduct](CODE_OF_CONDUCT.md).

## Testing

From the repository root (requires Go **1.25+**, same as [go.mod](go.mod)):

```bash
go test ./...
```

Stdio MCP end-to-end coverage (builds `cmd/service-manager-read-mcp-e2e`, then drives it over stdin/stdout with the official MCP client transport):

```bash
go test ./internal/e2e/...
```

For parity with CI:

```bash
go test -race ./...
```

## License

By contributing, you agree that your contributions are licensed under the Apache License 2.0, the same license as the project. See [LICENSE](LICENSE) and [NOTICE](NOTICE).
