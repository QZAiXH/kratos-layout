package job

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"
)

// HandlerFunc 表示后台任务处理函数。
type HandlerFunc func(context.Context, *asynq.Task) error

// Runtime 管理 Asynq worker 生命周期和任务路由。
type Runtime struct {
	enabled bool            // enabled 表示任务运行时是否启用。
	server  *asynq.Server   // server 是 Asynq worker 服务。
	mux     *asynq.ServeMux // mux 保存任务类型到处理器的映射。
	log     *slog.Logger    // log 是运行时日志器。
}

// NewRuntime 创建后台任务运行时。
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

// Handle 注册任务类型对应的处理器。
func (r *Runtime) Handle(taskType string, handler HandlerFunc) {
	if r == nil || r.mux == nil || handler == nil {
		return
	}
	r.mux.HandleFunc(taskType, handler)
}

// Start 启动后台任务 worker。
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

// Stop 关闭后台任务 worker。
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
