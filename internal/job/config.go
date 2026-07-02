package job

import (
	"time"

	"github.com/QZAiXH/kratos-layout/internal/conf"
)

const (
	defaultQueueName         = "default"
	defaultWorkerConcurrency = 4
	defaultTaskTimeout       = 30 * time.Minute
	defaultShutdownTimeout   = 15 * time.Second
)

type Config struct {
	Enabled           bool
	Queue             string
	WorkerConcurrency int
	TaskTimeout       time.Duration
	ShutdownTimeout   time.Duration
}

func NewConfig(cfg *conf.Job) *Config {
	config := &Config{
		Queue:             defaultQueueName,
		WorkerConcurrency: defaultWorkerConcurrency,
		TaskTimeout:       defaultTaskTimeout,
		ShutdownTimeout:   defaultShutdownTimeout,
	}
	if cfg == nil {
		return config
	}
	config.Enabled = cfg.GetEnabled()
	if cfg.GetQueue() != "" {
		config.Queue = cfg.GetQueue()
	}
	if cfg.GetWorkerConcurrency() > 0 {
		config.WorkerConcurrency = int(cfg.GetWorkerConcurrency())
	}
	if cfg.GetTaskTimeout() != nil && cfg.GetTaskTimeout().AsDuration() > 0 {
		config.TaskTimeout = cfg.GetTaskTimeout().AsDuration()
	}
	if cfg.GetShutdownTimeout() != nil && cfg.GetShutdownTimeout().AsDuration() > 0 {
		config.ShutdownTimeout = cfg.GetShutdownTimeout().AsDuration()
	}
	return config
}
