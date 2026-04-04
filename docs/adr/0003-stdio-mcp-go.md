# ADR 0003: stdio MCP server in Go

## Status

Accepted

## Context

The project targets IDE and agent clients that speak MCP over a subprocess with JSON-RPC on stdin/stdout.

## Decision

Implement the server with the official [Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk), **`StdioTransport`**, and structured tool outputs. Logs go to **stderr**; **stdout** is reserved for MCP.

## Consequences

- Clients configure a single binary path with empty `args` for normal operation.
- Go version follows SDK requirements (see `go.mod`).
