---
name: kratos-data-ent
description: Use for Ent schema changes, data repositories, transactions, Redis-backed stores, database access, and cross-layer mapping.
---

# Kratos Data And Ent

## Read First

- `internal/data/ent/schema`
- `internal/data/base`
- `internal/dep/db.go`
- Relevant `internal/data/<domain>`
- Relevant biz repo interface
- `internal/pkg/typecatch`
- `internal/pkg/pagination`

## Workflow

1. Define schemas in `internal/data/ent/schema`.
2. Prefer explicit table names with `entsql.Annotation` when schema ownership is stable.
3. Run `make ent`.
4. Keep repo interfaces in biz; implementations in data.
5. Translate Ent/Redis/system errors into domain errors before crossing upward.
6. Use DB constraints/transactions or Redis atomic operations for concurrency-sensitive state.

## Do Not

- Do not import Ent in biz or service.
- Do not hand-edit generated Ent files outside `schema`.
- Do not add in-memory locks for cross-instance coordination.

## Validation

```bash
make ent
go test ./internal/data/... ./internal/biz/...
```
