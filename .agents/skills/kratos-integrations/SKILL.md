---
name: kratos-integrations
description: Use for external HTTP clients, webhooks/callbacks, object storage, third-party APIs, Redis Streams/SSE, distributed locks, credentials, and integration tests.
---

# Kratos External Integrations

## Rules

- Put reusable external clients under `internal/pkg/<provider>` or `internal/dep` depending on lifecycle ownership.
- Put business-specific integration orchestration in biz.
- Never commit or print real credentials.
- Webhooks/callbacks must validate signatures, source, idempotency, amount/status semantics where applicable.
- Redis Streams/SSE state must be multi-instance safe.
- Use Redis/DB/object-storage primitives for cross-instance coordination.
- SSE handlers must set `Content-Type: text/event-stream`, `Cache-Control: no-cache`, `Connection: keep-alive`, and `X-Accel-Buffering: no`.
- SSE API docs must be backed by proto plus `docs/openapi/overlays/<module>.yaml`; include media type, headers, and representative frames.
- If runtime SSE is handwritten, do not also register the generated HTTP server for the same documented route.

## Validation

```bash
go test ./internal/pkg/... ./internal/dep/...
make openapi
```

Use e2e only when unit/integration tests cannot prove the behavior.
