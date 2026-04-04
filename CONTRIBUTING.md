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

Requires Go **1.25+** (see [go.mod](go.mod)).

**Full suite (matches CI intent):** install [golangci-lint](https://golangci-lint.run/welcome/install/), then from the repo root:

```bash
make ci
# or: ./scripts/ci.sh
```

That runs, in order: `gofmt` check, `go mod verify`, `go mod tidy` drift check, `go vet ./...`, `golangci-lint run`, `go build ./...`, unit tests with `-race -short`, then the stdio subprocess smoke test (`TestCLI_stdioSmoke` in `cmd/service-manager-read-mcp`).

**Focused commands:**

```bash
make test-short   # race + -short (skips stdio smoke)
make test-smoke   # only TestCLI_stdioSmoke
go test -race ./...
```

Automation and maintainer notes: [docs/agents.md](docs/agents.md).

## License

By contributing, you agree that your contributions are licensed under the Apache License 2.0, the same license as the project. See [LICENSE](LICENSE) and [NOTICE](NOTICE).
