package job

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

type Enqueuer struct {
	client *asynq.Client
	cfg    *Config
}

func NewEnqueuer(client *asynq.Client, cfg *Config) *Enqueuer {
	return &Enqueuer{client: client, cfg: cfg}
}

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

func UniqueTaskID(id string) asynq.Option {
	return asynq.TaskID(id)
}

func ProcessIn(delay time.Duration) asynq.Option {
	return asynq.ProcessIn(delay)
}
