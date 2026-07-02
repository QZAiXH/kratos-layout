package job

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

// Enqueuer 封装后台任务入队能力。
type Enqueuer struct {
	client *asynq.Client // client 是 Asynq 入队客户端。
	cfg    *Config       // cfg 是任务运行时配置。
}

// NewEnqueuer 创建后台任务入队器。
func NewEnqueuer(client *asynq.Client, cfg *Config) *Enqueuer {
	return &Enqueuer{client: client, cfg: cfg}
}

// EnqueueJSON 将 JSON payload 封装成任务并入队。
func (e *Enqueuer) EnqueueJSON(ctx context.Context, taskType string, payload any, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	if e == nil || e.client == nil || e.cfg == nil || !e.cfg.Enabled {
		return nil, fmt.Errorf("job runtime is disabled")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal task payload: %w", err)
	}
	options := []asynq.Option{asynq.Queue(e.cfg.Queue), asynq.Timeout(e.cfg.TaskTimeout)}
	options = append(options, opts...)
	return e.client.EnqueueContext(ctx, asynq.NewTask(taskType, body), options...)
}

// UniqueTaskID 返回唯一任务编号选项。
func UniqueTaskID(id string) asynq.Option {
	return asynq.TaskID(id)
}

// ProcessIn 返回延迟执行选项。
func ProcessIn(delay time.Duration) asynq.Option {
	return asynq.ProcessIn(delay)
}
