---
name: kratos-ent-data
description: Use when changing Ent schemas, generated Ent code, data repositories, transaction behavior, Redis-backed state, database constraints, or DTO mapping in a go-kratos service.
---

# Kratos Ent Data

Use this skill for Ent-backed persistence work in a go-kratos backend.

## Read First

- `internal/data/ent/schema` and generated Ent package layout.
- Data root, repo base, and transaction helpers under `internal/data`.
- Relevant `internal/data/<domain>` implementation.
- Relevant `internal/biz/<domain>` repo interfaces and use cases.
- Relevant `internal/dto/<domain>` definitions.
- Existing error packages, pagination helpers, lock helpers, and converter utilities.
- Project decisions about schema deletion, mixins, soft delete, and migration behavior.

## Workflow

1. Search for an existing schema, repo method, DTO converter, pagination helper, lock helper, or transaction helper before adding one.
2. Change Ent schemas under `internal/data/ent/schema` only. Use existing mixins and table-name annotations when the repo has them.
3. Encode idempotency and concurrency assumptions with DB constraints, indexes, transactions, or Redis primitives. Do not rely on process-local state for shared business data.
4. Regenerate Ent code with the repo command, usually `make ent`.
5. Keep repo interfaces in biz. Implement those interfaces in data.
6. Let biz own transaction boundaries. Data code should honor the transaction already carried by context.
7. Translate Ent, Redis, SQL, and integration errors into domain errors at the data boundary.
8. Convert persisted state to DTOs before crossing into biz/service layers.
9. Prefer existing copy helpers for same-name fields. Write explicit converters only for renames, normalization, or cross-field semantics.

## Guardrails

- Do not pass Ent entities into service output, proto responses, or biz contracts.
- Do not put repo interfaces in data if the architecture uses biz-owned ports.
- Do not bypass soft-delete behavior unless the path explicitly needs deleted rows.
- Do not add node-local caches for mutable shared business state.
- Do not use in-process locks for cross-request mutual exclusion in multi-instance deployments.
- Do not hand-edit generated Ent files outside schema sources.
- Do not add generic data helpers until at least two real call sites need them.

## Commands

Check the repo `Makefile` first. Common commands:

```bash
make ent
go test ./internal/data/<domain>/...
go test ./internal/biz/<domain>/...
go test -race ./internal/data/<domain>/...
golangci-lint run --fix
```

## Validation

- Verify constraints and indexes match the actual retry/concurrency behavior.
- Add the smallest test that fails if transaction, idempotency, soft-delete, or mapping behavior regresses.
- Run generation and build checks after schema or provider changes.
