---
name: kratos-reuse-map
description: Use before adding helpers, clients, converters, locks, pagination, errors, IDs, decimal utilities, or repeated module scaffolding.
---

# Kratos Reuse Map

## Existing Reuse Points

- `internal/pkg/typecatch`: copier-based same-name struct field copying.
- `internal/pkg/pagination`: page size/page number normalization.
- `internal/pkg/id`: prefixed ULID helpers. Entity and public IDs default to ULID strings.
- `internal/pkg/decimalx`: yuan/fen decimal helpers.
- `internal/pkg/token`: Ed25519 JWT token pair.
- `internal/pkg/authz`: current user extraction and JWT/Casbin middleware helpers.
- `internal/pkg/zaplog`: slog-compatible zap logger with masking and rotation.
- `internal/data/base`: shared data root.
- `internal/biz/base`: common usecase logger base.
- `internal/job`: Asynq runtime and enqueue helper.

## Typecatch

Use `typecatch.CopyTo[SRC, DST](&src)` only for boring same-name field mapping between adjacent layers, such as Ent entity to biz module result type. Keep explicit mapping when field names, units, permission filtering, or business meaning differ.

## Rule

Search existing helpers before adding new ones. Add a helper only when it removes real duplication or keeps layer boundaries cleaner.
