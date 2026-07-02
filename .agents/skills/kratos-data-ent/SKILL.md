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

1. Create a schema with `go run entgo.io/ent/cmd/ent@v0.14.6 new <Name> --target internal/data/ent/schema`, then edit the generated file.
2. All entity primary keys use `StringIDMixin{}` ULID string IDs. Use `StringIDMixin{Prefix: "xxx_"}` only when the prefix is part of the public contract.
3. Add `TimeMixin{}` to normal business tables. Add `SoftDeleteMixin{}` only when the table needs soft delete. Add `ByUserMixin{}` only for tables that must be created with authenticated user context.
4. Prefer explicit table names with `entsql.Annotation` when schema ownership is stable.
5. Keep `internal/data/ent/template/database.tmpl` wired through `internal/data/ent/generate.go`; it generates `ent.Database` with `InTx`, `GetClient`, `Exec`, `Query`, and ctx-aware entity clients.
6. Run `make ent`.
7. Keep repo interfaces in biz; implementations in data.
8. Use `*ent.Database` in data dependencies, not raw `*ent.Client`; call `db.GetClient()` only when raw Ent client access is required.
9. Translate Ent/Redis/system errors into domain errors before crossing upward.
10. Use DB constraints/transactions or Redis atomic operations for concurrency-sensitive state.
11. Map Ent entities to `internal/biz/<module>/types.go` types before returning. Use `typecatch.CopyTo[SRC, DST](&src)` only when same-name fields mean the same thing; otherwise map explicitly.

## Do Not

- Do not import Ent in biz or service.
- Do not hand-edit generated Ent files outside `schema`.
- Do not remove the Ent database template from generation when changing schema code.
- Do not add in-memory locks for cross-instance coordination.
- Do not return Ent/Redis/SQL errors directly to biz/service.
- Do not create `internal/dto`; map data results to structs owned by the consuming biz module, starting in `types.go` and splitting by purpose inside that module when needed.

## Validation

```bash
go run entgo.io/ent/cmd/ent@v0.14.6 new Todo --target internal/data/ent/schema
make ent
go test ./internal/data/... ./internal/biz/...
```
