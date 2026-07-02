---
name: kratos-auth-permission
description: Use when changing JWT authentication, token handling, Casbin authorization, anonymous whitelist, roles, policies, or security-sensitive endpoint access.
---

# Kratos Auth And Permission

## Read First

- `internal/server/security.go`
- `internal/pkg/authz`
- `internal/pkg/token`
- `internal/dep/casbin.go`
- Generated operation constants under `api/**`

## Mental Model

- JWT middleware validates Ed25519 access tokens.
- Casbin authorizes `(subject, operation, action)`.
- Anonymous routes are controlled by `NewProtectedMatcher`.
- Template Todo APIs are anonymous examples; new business APIs should be protected by default.

## Rules

- Whitelist generated operation constants, not URL prefixes.
- Do not put roles in JWT as the authorization source of truth.
- Do not expose access/refresh tokens outside auth responses.
- Callback/webhook routes may be anonymous only if payload/signature validation exists.

## Validation

```bash
go test ./internal/server ./internal/pkg/authz ./internal/pkg/token ./internal/dep
```
