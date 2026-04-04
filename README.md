# service-manager-read-mcp

A **stdio** [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) server that exposes **read-only**, **structured** access to service state on Linux and macOS.

The server is **safe by default**: it does not start, stop, or reconfigure services. It wraps allowlisted supervisor CLIs so agents can query status and metadata in a consistent shape across platforms.

**Threat model (skim):** the process only runs **known binaries** with **fixed argv templates** and **validated parameter slots**—no shell, no free-form flags from tool input, bounded stdout/stderr and timeouts. It is aimed at **read-only inspection** on a machine where you already trust running this binary; it does not sandbox the host against a malicious local user. For boundaries and layering, see [docs/architecture.md](docs/architecture.md).

## Documentation

All detailed documentation lives under **`docs/`**. Start at **[docs/README.md](docs/README.md)**.

| | |
| --- | --- |
| MCP tools (table, examples, build/CI summary) | [docs/mcp-tools.md](docs/mcp-tools.md) |
| Install, client config, `go install` | [docs/distribution.md](docs/distribution.md) |
| Supervisors: scope, detection, backends | [docs/supervisors.md](docs/supervisors.md) |
| Repo layout & local dev | [docs/repository-layout.md](docs/repository-layout.md) |
| Architecture & safety model | [docs/architecture.md](docs/architecture.md) |
| Changelog | [docs/CHANGELOG.md](docs/CHANGELOG.md) |
| Releases & tagging | [docs/releasing.md](docs/releasing.md) |
| Contributing & tests | [docs/contributing.md](docs/contributing.md) |
| Security | [docs/security.md](docs/security.md) |
| ADRs | [docs/adr/README.md](docs/adr/README.md) |

## License

Apache License 2.0. See [LICENSE](LICENSE) and [NOTICE](NOTICE).

## Contributing

See [docs/contributing.md](docs/contributing.md). [Code of Conduct](docs/code-of-conduct.md) · [Security reporting](docs/security.md).
