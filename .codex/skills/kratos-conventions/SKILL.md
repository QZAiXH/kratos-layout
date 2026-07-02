---
name: kratos-conventions
description: Lightweight entrypoint for this Kratos template. Use before code or project-file modifications, especially API/proto/schema/Wire/test/devops changes.
---

# Kratos Template Conventions

Read `AGENTS.md`, then load the smallest matching `.agents/skills/kratos-*` skill:

- `kratos-architecture-navigation`
- `kratos-backend-api`
- `kratos-data-ent`
- `kratos-auth-permission`
- `kratos-job-runtime`
- `kratos-integrations`
- `kratos-test-quality`
- `kratos-devops`
- `kratos-reuse-map`

Keep this entrypoint thin; update `AGENTS.md` or the specific skill when project rules change.
