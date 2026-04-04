# Distribution and installation

This project ships as a **single static Go binary** (`service-manager-read-mcp`). There is no official OS package manager yet; installation is **from source** with Go **1.25+** (see [go.mod](../go.mod)).

## What you are installing

- A **stdio MCP server**: the process speaks [Model Context Protocol](https://modelcontextprotocol.io/) JSON-RPC on **stdin/stdout**. **Do not** attach a terminal to stdout or pipe unrelated data there.
- **stderr** is reserved for server logs (text). MCP clients should still spawn the process with stderr available (or discard it); protocol traffic never uses stderr.
- **Configuration**: none required. Discovery uses `PATH` and `GOOS` only (see [README](../README.md) and [architecture.md](architecture.md)).
- **Supervisor CLIs** (`systemctl`, `launchctl`, etc.) must exist on the host for the corresponding backends to work; the server does not bundle them.

## Install with Go

### Latest module version (tagged releases)

When a version tag exists:

```bash
go install github.com/ipolyzos/service-manager-read-mcp/cmd/service-manager-read-mcp@latest
```

The installed binary’s MCP `implementation.version` comes from the **module version** (e.g. `v0.4.0`). Use `service-manager-read-mcp -version` for commit/build metadata when the binary was built with module info (`-buildvcs` defaults help).

### Clone and build

```bash
git clone https://github.com/ipolyzos/service-manager-read-mcp.git
cd service-manager-read-mcp
make build
# binary: bin/service-manager-read-mcp (gitignored)
```

### Release-style binary (embedded semver)

For releases or local packaging, set an explicit semver and embed commit/time:

```bash
make build-release VERSION=0.4.0
./bin/service-manager-read-mcp -version
```

Variables (see [Makefile](../Makefile)):

| Makefile var | `internal/version` field | Purpose |
| ------------ | ------------------------ | ------- |
| `VERSION` | `Version` | Release semver (optional; empty uses module/`0.0.0-dev` logic) |
| `COMMIT` | `Commit` | Short SHA (default: `git rev-parse --short HEAD`) |
| `BUILD_DATE` | `BuildDate` | UTC timestamp (default: `date -u …`) |

Equivalent raw `go build`:

```bash
go build -trimpath -ldflags "\
  -X 'github.com/ipolyzos/service-manager-read-mcp/internal/version.Version=0.4.0' \
  -X 'github.com/ipolyzos/service-manager-read-mcp/internal/version.Commit=abc123' \
  -X 'github.com/ipolyzos/service-manager-read-mcp/internal/version.BuildDate=2026-04-04T12:00:00Z'" \
  -o service-manager-read-mcp ./cmd/service-manager-read-mcp
```

## CLI

| Flag | Meaning |
| ---- | ------- |
| `-version` | Print version line and exit (does not start MCP). |
| `-h` / `-help` | Standard Go flag help. |

**MCP mode:** run with **no flags** (typical). Clients pass **`args`: `[]`** so stdin/stdout stay dedicated to JSON-RPC.

## MCP client configuration

Use an **absolute path** to the binary. Example (shape varies by client):

```json
{
  "mcpServers": {
    "service-manager-read": {
      "command": "/usr/local/bin/service-manager-read-mcp",
      "args": []
    }
  }
}
```

## Release expectations (maintainers)

- Tag **semver** releases on the integration branch policy the repo uses (see [CONTRIBUTING.md](../CONTRIBUTING.md)).
- Prefer **`make build-release VERSION=x.y.z`** (or CI equivalent) so binaries report a clear semver and commit.
- Publishing **module** artifacts is via Git tags (`go install …@vx.y.z`); optional GitHub **Releases** with attached binaries are a separate step if the project adopts them later.

## Related docs

- [README.md](../README.md) — tools, platforms, detection table  
- [architecture.md](architecture.md) — layering and safety model  
