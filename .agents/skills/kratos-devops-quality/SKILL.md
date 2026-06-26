---
name: kratos-devops-quality
description: Use when working on go-kratos local startup, Makefile targets, code generation, config proto generation, OpenAPI generation, lint, tests, Docker images, CI, or deployment troubleshooting.
---

# Kratos DevOps Quality

Use this skill for build, generation, verification, and deployment work in a go-kratos backend.

## Read First

- `README.md`, `Makefile`, `go.mod`, and tool installation docs.
- Config proto files, generated config files, and example config files.
- API generation and OpenAPI docs.
- Dockerfile, compose files, deployment scripts, and CI configuration.
- Test docs and e2e prerequisites.
- Real config files for field names only. Do not copy or quote secret values.

## Workflow

1. Inspect the repo `Makefile` before inventing commands.
2. Match source changes to generation:
   - API proto -> API generation, HTTP/gRPC bindings, validation, OpenAPI.
   - Config proto -> config generation and config examples.
   - Ent schema -> Ent generation.
   - Wire providers -> Wire generation.
3. After Go code changes, run the repo lint command, usually `golangci-lint run --fix`, and require zero remaining issues.
4. Run the smallest test package set that proves the change. Broaden to `go test ./...` only for shared behavior.
5. Keep e2e tests opt-in unless the changed behavior cannot be proven without real external services.
6. For Docker/deploy changes, verify config mount paths, required env vars, runtime image assumptions, and tag/release gates.

## Guardrails

- Do not print secrets from local config, CI variables, deploy env files, or logs.
- Do not hand-edit generated OpenAPI, generated protobuf, generated config, generated Wire, or generated Ent files.
- Do not treat example env files as complete runtime config.
- Do not assume a minimal production image has a shell or debugging tools.
- Do not skip lint because the code change is small.
- Do not add a new tool or dependency when an existing Make target or Go toolchain command already does the job.

## Commands

Check names in the repo first. Common commands:

```bash
make init
make api
make config
make ent
make openapi
make generate
make build
golangci-lint run --fix
go test ./...
go run ./cmd/<app> -conf ./configs
docker build -t <image> .
```

## Validation

- Config changes: regenerate config, inspect example config diffs, and run targeted startup/config tests.
- Deploy scripts: run shell syntax checks and review required variables.
- Docker changes: build the image when feasible.
- Documentation-only changes: inspect diff and verify paths/commands still exist.
