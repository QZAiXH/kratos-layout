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
- `internal/biz/biz.go`
- `internal/data/data.go`
- `internal/service/service.go`
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

- Keep tiny modules flat: `internal/biz/todo.go`, `internal/data/todo.go`, `internal/service/todo.go`.
- Split a growing module into matching directories: `internal/biz/<module>`, `internal/data/<module>`, `internal/service/<module>`.
- In `internal/biz/<module>`, use:
  - `use_case.go` for `UseCase`, `Repo`, narrow cross-module Provider interfaces, constructors.
  - `types.go` for module request/result/value types and status constants. If it grows, split by purpose inside the same module, such as `request.go`, `result.go`, `model.go`, or `status.go`.
  - action files such as `command.go`, `query.go`, `validate.go`, `stream.go`, `job.go` as the module grows.
- Put cross-module Provider interfaces in the consuming biz module. Bind implementations in the top-level provider set.

## Do Not

- Do not pass Ent entities across layer boundaries.
- Do not put business logic in service handlers.
- Do not hand-edit generated files.
- Do not add a central DTO package; keep shared structs in the owning biz module.
- Do not add cloud/vendor-specific code to the main template unless it is required by most generated projects.

## Validation

```bash
make all
go test ./...
make build
```
