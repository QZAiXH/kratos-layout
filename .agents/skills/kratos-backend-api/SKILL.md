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
- Use generated operation constants for auth whitelist/policy decisions.
- Do not hand-write normal HTTP business routes when proto generation covers the route.
- For handwritten SSE routes, keep a document-only proto if useful for OpenAPI, but do not register the generated HTTP server for that proto path.
- Document SSE responses through `docs/openapi/overlays/<module>.yaml` with `text/event-stream`, stream headers, and examples.
- Put enum and enum value descriptions in proto comments; the OpenAPI publisher emits `x-enum-varnames`, `x-enum-descriptions`, and an enum description block.
- Do not expose secrets, tokens, passwords, or internal credentials in responses.

## Validation

```bash
make api
make openapi
make generate
go test ./...
make build
```
