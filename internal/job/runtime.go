package job

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"
)

type HandlerFunc func(context.Context, *asynq.Task) error

type Runtime struct {
	enabled bool
	server  *asynq.Server
	mux     *asynq.ServeMux
	log     *slog.Logger
}

func NewRuntime(cfg *Config, redisOpt asynq.RedisConnOpt, logger *slog.Logger) *Runtime {
	runtime := &Runtime{log: logger}
	if cfg == nil || !cfg.Enabled || redisOpt == nil {
		return runtime
	}
	runtime.enabled = true
	runtime.mux = asynq.NewServeMux()
	runtime.server = asynq.NewServer(redisOpt, asynq.Config{
		Concurrency: cfg.WorkerConcurrency,
		Queues:      map[string]int{cfg.Queue: 1},
	})
	return runtime
}

func (r *Runtime) Handle(taskType string, handler HandlerFunc) {
	if r == nil || r.mux == nil || handler == nil {
		return
	}
	r.mux.HandleFunc(taskType, handler)
}

func (r *Runtime) Start(ctx context.Context) error {
	if r == nil || !r.enabled || r.server == nil || r.mux == nil {
		return nil
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- r.server.Run(r.mux)
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return fmt.Errorf("start asynq worker: %w", err)
	default:
		if r.log != nil {
			r.log.Info("job runtime started")
		}
		return nil
	}
}

func (r *Runtime) Stop(context.Context) error {
	if r == nil || !r.enabled || r.server == nil {
		return nil
	}
	r.server.Shutdown()
	if r.log != nil {
		r.log.Info("job runtime stopped")
	}
	return nil
}
