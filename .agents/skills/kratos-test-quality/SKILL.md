---
name: kratos-test-quality
description: Use for tests, lint failures, generated-code validation, route/auth tests, job tests, data tests, or final verification strategy.
---

# Kratos Test And Quality

## Workflow

1. Pick the smallest package set that proves the change.
2. Regenerate code before testing generated consumers.
3. Run `make openapi` when proto comments, enums, SSE overlays, or public routes change.
4. Use `go test ./...` for shared infrastructure changes.
5. Use race tests for lock/concurrency-sensitive code.
6. Run lint before publishing a template change.

## Comment Rules

- Every handwritten Go function, method, interface, interface method, struct, and struct field needs a Chinese comment.
- Every top-level `Test*` and every `t.Run` sub-case needs a Chinese comment describing business intent, precondition, and expected behavior.
- `go test ./...` runs `internal/pkg/commentcheck`, which skips generated files and checks handwritten Go comments.

## Commands

```bash
go test ./...
make openapi
go test -race ./internal/job ./internal/dep
make build
make lint
```

E2E tests should be opt-in and documented with Docker/Redis/Postgres prerequisites.
