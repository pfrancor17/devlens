# ADR-002: Hexagonal architecture for the Go server

- **Status:** Accepted
- **Date:** 2026-10-26

## Context

DevLens will talk to several external systems that are likely to change: the GitHub API (and possibly other providers later), an ONNX runtime, a result store (SQLite first, maybe Postgres) and a vector index (still undecided). The business rules (how risk is interpreted, what counts as a duplicate, how a plan is compared to a PR) must stay stable and testable while those pieces change.

## Decision

The server follows a hexagonal (ports and adapters) architecture:

| Layer | Package | Depends on |
| --- | --- | --- |
| Domain | `internal/domain` | Standard library only |
| Use cases | `internal/usecase/<module>` | Domain, ports |
| Driven ports | `internal/port` | Domain |
| Driven adapters | `internal/adapter/<name>` | Ports, external libraries |
| Driving adapter (MCP) | `internal/mcpserver` | Use cases, MCP SDK |
| Composition root | `cmd/devlens` | Everything; wires it together |

Rules:

- Dependencies point inward only. The domain never imports another layer.
- A test (`internal/domain/architecture_test.go`) fails the build if the domain imports anything outside the standard library.
- Consumer-defined interfaces: a port is declared where it is used, and only when a use case needs it.
- The MCP layer translates requests into use-case calls and contains no business rules.

## Consequences

**Positive**

- Use cases are tested with fake adapters: fast tests with no network or model files.
- Swapping SQLite for Postgres, or adding GitLab, touches one adapter.
- The structure itself documents where each kind of code belongs.

**Negative**

- More packages and indirection than a small script would need. Accepted: the project is expected to grow to five modules.
- Discipline is required to keep rules out of adapters. Enforced through code review and, from phase 2, by DevLens's own Senior Reviewer.
