# Repository layout and development

## Layout

```
.github/
  dependabot.yml                # weekly PRs: Go modules + GitHub Actions
  workflows/
    ci.yml                      # push/PR → reusable ci-verify
    ci-verify.yml               # shared gofmt, modules, vet, lint, build, test
    release.yml                 # tag v* → ci-verify, then GitHub Release
scripts/
  ci.sh                         # local parity with ci-verify (toolchain + lint checks)
  golangci-version              # single line: golangci-lint version (must match CI)
cmd/service-manager-read-mcp/   # MCP stdio entrypoint (+ integration smoke test)
cmd/service-manager-read-mcp-e2e/ # test-only stdio server (deterministic fake backend)
internal/
  e2e/                          # MCP stdio e2e tests (subprocess + CommandTransport)
  servicemcp/                   # MCP tool schemas, handlers, RegisterTools, RunStdio
  servicemcp/e2etest/           # deterministic fake backend for e2e binary
  version/                      # binary & MCP implementation version metadata
  domain/                       # shared types (units, status, logs, capabilities)
  svcerr/                       # typed errors for backends and runners
  runner/                       # argv-only subprocess runner (+ test fake)
  detect/                       # PATH-based supervisor discovery (no subprocess)
  backend/                      # ServiceManagerBackend interface
  backend/stub/                 # noop backend for wiring tests
  backend/systemd/              # systemd/journald backend (allowlisted CLIs)
  backend/launchd/              # launchd backend (allowlisted launchctl)
  backend/openrc/               # OpenRC backend (rc-service / rc-status)
  backend/runit/                # runit backend (allowlisted sv)
  backend/brew/                 # Homebrew services backend (allowlisted brew)
```

## Local development

Requires [Go 1.25+](https://go.dev/dl/) for the MCP SDK.

```bash
make build
make build-all
go test ./...
go vet ./...
```

Subprocess execution for backends is routed through `internal/runner`: build an allowlisted `runner.Registry`, wrap a `runner.Exec` (or test `runner.Func`) with `runner.SafeRunner`, and call `RunInvocation`—never hand-construct argv from tool or agent input. MCP tools call `internal/servicemcp` only, which uses `detect` and `backend` packages. See [architecture.md](architecture.md).

More testing targets and the host smoke driver: **[contributing.md](contributing.md)**.
