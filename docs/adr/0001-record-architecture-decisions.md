# ADR 0001: Record architecture decisions

## Status

Accepted

## Context

The project spans MCP, multiple OS supervisors, and a strict execution model. Informal decisions are easy to lose across issues and PRs.

## Decision

Use numbered markdown ADRs under `docs/adr/` for decisions that are hard to reverse or that explain **why** the code is shaped as it is.

## Consequences

- Contributors can discover intent without archaeology in old threads.
- `docs/architecture.md` stays a concise map; ADRs hold durable rationale.
