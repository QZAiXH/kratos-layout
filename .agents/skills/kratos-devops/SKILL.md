---
name: kratos-devops
description: Use for Makefile, toolchain setup, config generation, Docker, docker-compose, local startup, OpenAPI generation, release, or deploy troubleshooting.
---

# Kratos DevOps

## Read First

- `Makefile`
- `go.mod`
- `Dockerfile`
- `deploy/docker-compose.yml`
- `internal/conf/conf.proto`
- `configs/config.yaml`
- `scripts/openapi`
- `docs/openapi`

## Facts

- `make build` auto-detects the single directory under `cmd/`.
- Template repo keeps `cmd/server`; generated repos get `cmd/<project>`.
- Runtime config path is passed with `-conf`.
- Docker runtime reads config from `/data/conf`.
- `make openapi` generates OpenAPI 3.1 module docs and `docs/openapi/bundles/openapi.json`.
- OpenAPI generation uses `buf build/generate`; `protoc-gen-openapi v0.7.1` is invoked through `go run` for the baseline.

## Validation

```bash
make all
make openapi
make build
go test ./...
docker build -t kratos-template .
```
