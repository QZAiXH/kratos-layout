---
name: kratos-backend-api
description: Use when adding or changing protobuf contracts, generated HTTP/gRPC APIs, service handlers, biz use cases, data repos, error definitions, or OpenAPI output.
---

# Kratos Backend API

## Workflow

1. Change proto first under `api/<domain>/v1` or `api/<domain>/<api>/v1`.
2. Run `make api`.
3. Run `make openapi` when the proto affects public HTTP docs.
4. Implement service as conversion and error propagation only.
5. Put validation, orchestration, idempotency, and transaction boundaries in biz.
6. Put storage access in data.
7. Register new generated service in `internal/server/http.go` and `internal/server/grpc.go`.
8. Update provider sets and run `make generate`.

## Rules

- Proto is the source of truth.
- Proto services, RPCs, messages, fields, enums, and enum values must have Chinese comments because they feed the OpenAPI publisher.
- For a non-trivial module, put shared request/result structs and status constants in `internal/biz/<module>/types.go`; if it grows, split by purpose inside the same module. Do not add `internal/dto`.
- Keep only `UseCase`, dependency fields, `Repo`, Provider interfaces, and constructors in `internal/biz/<module>/use_case.go`; put use case methods in `command.go`, `query.go`, `stream.go`, or `job.go`.
- Keep only `Service`, generated server embedding, and `NewService` in `internal/service/<module>/service.go`; put RPC methods in `command.go`, `query.go`, or `stream.go`, and proto/biz conversion in `mapper.go`.
- Keep top-level `internal/{biz,data,service}/provider.go` files limited to Wire provider sets, binds, simple aliases, and imports.
- Handwritten Go functions, methods, interfaces, interface methods, structs, and every struct field must have Chinese comments.
- Use generated operation constants for auth whitelist/policy decisions.
- Do not hand-write normal HTTP business routes when proto generation covers the route.
- For handwritten SSE routes, keep a document-only proto if useful for OpenAPI, but do not register the generated HTTP server for that proto path.
- Document SSE responses through `docs/openapi/overlays/<module>.yaml` with `text/event-stream`, stream headers, and examples.
- Put enum and enum value descriptions in proto comments; the OpenAPI publisher emits `x-enum-varnames`, `x-enum-descriptions`, and an enum description block.
- Define public API errors in module error proto files first. Return those Kratos errors from biz/data and let service pass them through.
- Use private sentinel errors only for package-local control flow or Provider contracts; keep them in module `errors.go` when needed.
- Do not expose secrets, tokens, passwords, or internal credentials in responses.

## Validation

```bash
make api
make openapi
make generate
go test ./...
make build
```
