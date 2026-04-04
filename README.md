# service-manager-read-mcp

A **stdio** [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) server that exposes **read-only**, **structured** access to service state on Linux and macOS.

The server is designed to be **safe by default**: it does not start, stop, or reconfigure services. It wraps common service-manager command-line interfaces so agents and tools can query status and metadata in a consistent shape across platforms.

## Scope (planned)

Platform tooling this project intends to wrap (where available on the host):

| Platform / ecosystem | Typical CLI |
| ---------------------- | ----------- |
| systemd (Linux)      | `systemctl` |
| launchd (macOS)      | `launchctl` |
| OpenRC               | `rc-service` |
| runit                | `sv`        |
| Homebrew (macOS)     | `brew services` |

## Project layout

```
cmd/service-manager-read-mcp/   # process entrypoint (MCP wiring pending)
internal/
  domain/                       # shared types (units, status, logs, capabilities)
  svcerr/                       # typed errors for backends and runners
  runner/                       # argv-only subprocess runner (+ test fake)
  backend/                      # ServiceManagerBackend interface
  backend/stub/                 # noop backend for wiring tests
```

## Development

Requires [Go 1.22+](https://go.dev/dl/).

```bash
go build -o bin/service-manager-read-mcp ./cmd/service-manager-read-mcp
go test ./...
go vet ./...
```

## Status

Domain model, backend interface, subprocess runner abstraction, and a stub backend are in place. MCP stdio wiring and real supervisor integrations are not implemented yet.

## License

Apache License 2.0. See [LICENSE](LICENSE) and [NOTICE](NOTICE).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). The project uses a [Code of Conduct](CODE_OF_CONDUCT.md).
