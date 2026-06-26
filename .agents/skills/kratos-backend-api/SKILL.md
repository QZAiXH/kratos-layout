---
name: kratos-backend-api
description: Use when adding or changing go-kratos backend API contracts, protobuf HTTP/gRPC endpoints, service handlers, biz use cases, DTOs, error contracts, route registration, or Wire provider wiring across service, biz, and data layers.
---

# Kratos Backend API

Use this skill for proto-first API work in a go-kratos backend.

## Read First

- Project `AGENTS.md` and any architecture notes.
- Relevant `api/**/v1/*.proto` files and error proto files.
- Relevant `internal/service/<domain>`, `internal/biz/<domain>`, `internal/data/<domain>`, and `internal/dto/<domain>`.
- Route registration in `internal/server`.
- Provider sets and top-level Wire entrypoints when constructors or dependencies change.
- OpenAPI generation docs if the project publishes generated API documents.

## Workflow

1. Search existing endpoints, DTOs, errors, converters, and route patterns before adding code.
2. Treat proto as the contract source. Change RPCs, messages, HTTP annotations, validate rules, and error definitions first.
3. Regenerate API code with the repo command, usually `make api`.
4. Add or update DTOs for cross-layer data. Do not expose generated proto or storage entities as biz/data contracts unless the project already does.
5. Keep service handlers thin: proto/DTO conversion, current-user extraction, logging context, and error propagation.
6. Put validation, orchestration, idempotency, authorization calls, and transaction boundaries in biz use cases.
7. Implement persistence behind repo interfaces. Translate storage and integration errors into domain errors at the boundary.
8. Register generated HTTP/gRPC services in server setup when adding a new API surface.
9. Update ProviderSets and rerun Wire only when new constructors or dependencies are introduced.

## Guardrails

- Do not put business rules in service handlers.
- Do not import Ent, SQL clients, Redis clients, or concrete storage into biz unless the local architecture explicitly allows it.
- Do not let Ent entities cross into service/API output.
- Do not hand-edit generated `.pb.go`, `*_http.pb.go`, `*_grpc.pb.go`, `*.pb.validate.go`, generated error files, or generated OpenAPI JSON.
- Do not add a shared abstraction for one endpoint. Reuse local helpers first; otherwise write the smallest local code that proves the behavior.
- Do not expose secrets, tokens, credentials, or internal config values in API responses, logs, examples, or tests.

## Commands

Check the repo `Makefile` first. Common commands:

```bash
make api
make openapi
make generate
wire ./...
go build ./...
golangci-lint run --fix
go test ./...
```

## Validation

- Check route registration tests when HTTP paths, anonymous access, middleware behavior, or generated operation names change.
- Check OpenAPI output only through generation or overlays, not manual edits.
- Run targeted service/biz/data tests for the touched domain; broaden only when shared behavior changed.
