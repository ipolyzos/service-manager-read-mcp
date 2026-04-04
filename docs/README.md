# Documentation

Canonical documentation for **service-manager-read-mcp** lives in this directory. The root [README.md](../README.md) is a short entrypoint with links here only.

## Start here

| Topic | Document |
| ----- | -------- |
| **MCP tools**, stdio rules, tool table, build/install summary | [mcp-tools.md](mcp-tools.md) |
| Install, MCP clients, `go install`, release-style builds | [distribution.md](distribution.md) |
| Scope, PATH detection, systemd / launchd / OpenRC details | [supervisors.md](supervisors.md) |
| Repository tree, local dev commands | [repository-layout.md](repository-layout.md) |
| Architecture & backend contract | [architecture.md](architecture.md) |
| “Failed” semantics per supervisor | [failed-services.md](failed-services.md) |
| Recent logs / journal limits | [recent-logs.md](recent-logs.md) |

## Project operations

| Topic | Document |
| ----- | -------- |
| Branch flow, CI vs local, release prep context | [agents.md](agents.md) |
| Semver tags, `go install @vX.Y.Z`, tag → GitHub Release workflow | [releasing.md](releasing.md) |
| Dependabot (weekly Go + Actions PRs) | [`.github/dependabot.yml`](../.github/dependabot.yml) (config only) |
| Version history | [CHANGELOG.md](CHANGELOG.md) |
| Architecture decision records | [adr/README.md](adr/README.md) |

## Community

| Topic | Document |
| ----- | -------- |
| Contributing workflow, testing, host smoke script | [contributing.md](contributing.md) |
| Security reporting | [security.md](security.md) |
| Code of Conduct | [code-of-conduct.md](code-of-conduct.md) |

Root copies of **CONTRIBUTING**, **SECURITY**, **CHANGELOG**, and **CODE_OF_CONDUCT** are short stubs so GitHub and old links keep working; the full text is under **docs/**.
