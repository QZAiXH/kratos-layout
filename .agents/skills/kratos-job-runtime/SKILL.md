---
name: kratos-job-runtime
description: Use when adding, changing, debugging, or testing background jobs, Asynq enqueueing, worker behavior, queues, retries, or Redsync/distributed job locks.
---

# Kratos Job Runtime

## Read First

- `internal/job`
- `internal/dep/job.go`
- `internal/conf/conf.proto`
- `configs/config.yaml`

## Workflow

1. Keep job runtime config in `conf.Job`.
2. Use `internal/job.Enqueuer` for ad-hoc enqueueing.
3. Register handlers from application code; keep handlers thin.
4. Put business transitions and repo calls in biz.
5. Use Redis/Redsync or DB constraints for distributed mutual exclusion.

## Do Not

- Do not add per-domain goroutine/ticker loops.
- Do not call data repos directly from generic job runtime code.
- Do not treat enqueue success as business completion.

## Validation

```bash
go test ./internal/job ./internal/dep
```
