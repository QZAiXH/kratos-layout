---
name: kratos-reuse-map
description: Use before adding helpers, clients, converters, locks, pagination, errors, IDs, decimal utilities, or repeated module scaffolding.
---

# Kratos Reuse Map

## Existing Reuse Points

- `internal/pkg/typecatch`: copier-based same-name struct field copying.
- `internal/pkg/pagination`: page size/page number normalization.
- `internal/pkg/id`: UUID and prefixed ULID helpers.
- `internal/pkg/decimalx`: yuan/fen decimal helpers.
- `internal/pkg/token`: Ed25519 JWT token pair.
- `internal/pkg/authz`: current user extraction and JWT/Casbin middleware helpers.
- `internal/pkg/zaplog`: slog-compatible zap logger with masking and rotation.
- `internal/data/base`: shared data root.
- `internal/biz/base`: common usecase logger base.
- `internal/job`: Asynq runtime and enqueue helper.

## Rule

Search existing helpers before adding new ones. Add a helper only when it removes real duplication or keeps layer boundaries cleaner.
