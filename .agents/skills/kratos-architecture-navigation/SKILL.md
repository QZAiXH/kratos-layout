---
name: kratos-architecture-navigation
description: Use before non-trivial Kratos template work, feature design, issue analysis, refactors, module expansion, provider changes, generated-code workflows, or unclear architecture boundaries.
---

# Kratos Architecture Navigation

## Read First

- `AGENTS.md`
- `cmd/server/main.go`
- `cmd/server/wire.go`
- `internal/server/http.go`
- `internal/server/grpc.go`
- `internal/biz/provider.go`
- `internal/data/provider.go`
- `internal/service/provider.go`
- `internal/dep/provider.go`
- `internal/job` when background work is involved

## Workflow

1. Identify the touched layer: `api`, `service`, `biz`, `data`, `dep`, `server`, `job`, or `pkg`.
2. Keep request flow `service -> biz -> data`.
3. Keep external dependency construction in `internal/dep`.
4. Update provider sets and regenerate Wire when constructors change.
5. Route background execution through `internal/job`; do not add local goroutine/ticker runtimes.
6. Preserve template compatibility: `cmd/server` in the template repo, no project-name hardcoding.

## Module Layout

- Keep modules in matching directories even when small: `internal/biz/<module>`, `internal/data/<module>`, `internal/service/<module>`.
- In `internal/biz/<module>`, use:
  - `use_case.go` only for `UseCase`, dependency fields, `Repo`, narrow cross-module Provider interfaces, and constructors. Do not add business method bodies there.
  - `types.go` for module request/result/value types and status constants. If it grows, split by purpose inside the same module, such as `request.go`, `result.go`, `model.go`, or `status.go`.
  - action files such as `command.go`, `query.go`, `validate.go`, `stream.go`, `job.go` as the module grows.
- In `internal/data/<module>`, use `repo.go` only for the repo struct, dependency fields, and `NewRepo`. Put repo methods in `command.go` and `query.go`; put Ent/Redis details in `store.go`, `cache.go`, or `mapper.go`.
- In `internal/service/<module>`, use `service.go` only for the service struct and `NewService`. Put RPC methods in `command.go`, `query.go`, or `stream.go`; put proto/biz conversion in `mapper.go`.
- Top-level `internal/{biz,data,service}/provider.go` files only aggregate Wire provider sets, binds, simple aliases, and imports.
- Put cross-module Provider interfaces in the consuming biz module. Bind implementations in the top-level provider set.

## Do Not

- Do not pass Ent entities across layer boundaries.
- Do not put business logic in service handlers.
- Do not add CRUD, query, stream, mapping, validation, or transaction logic to `provider.go`, `use_case.go`, `repo.go`, or `service.go` entry files.
- Do not hand-edit generated files.
- Do not add a central DTO package; keep shared structs in the owning biz module.
- Do not add cloud/vendor-specific code to the main template unless it is required by most generated projects.

## Comment Rules

- Handwritten Go functions, methods, interfaces, interface methods, structs, and every struct field must have Chinese comments.
- Generated files are excluded from manual comment work.

## Validation

```bash
make all
go test ./...
make build
```
