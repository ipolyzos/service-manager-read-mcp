# Security

## Supported versions

Security fixes are applied on the current **`develop`** integration branch and released from **`main`** via semver tags (see [docs/releasing.md](docs/releasing.md)).

## Reporting a vulnerability

**Do not** open a public GitHub issue for undisclosed security vulnerabilities.

Please report sensitive issues using **[GitHub private vulnerability reporting](https://github.com/ipolyzos/service-manager-read-mcp/security/advisories/new)** for this repository (if enabled), or contact the maintainers privately with enough detail to reproduce and assess impact.

Include affected version or commit, component (e.g. MCP layer vs a specific backend), and any suggested fix if you have one.

## Scope notes

This project runs on the user’s machine and invokes **allowlisted** supervisor CLIs only. Reports about “running as the same user as the IDE” are generally expected; we still want to hear about **unexpected** command execution, argv injection, or unsafe defaults.
