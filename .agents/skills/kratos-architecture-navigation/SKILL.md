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

## Do Not

- Do not pass Ent entities across layer boundaries.
- Do not put business logic in service handlers.
- Do not hand-edit generated files.
- Do not add cloud/vendor-specific code to the main template unless it is required by most generated projects.

## Validation

```bash
make all
go test ./...
make build
```
