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

// Config 表示后台任务运行时配置。
type Config struct {
	Enabled           bool          // Enabled 表示是否启用后台任务运行时。
	Queue             string        // Queue 是默认队列名称。
	WorkerConcurrency int           // WorkerConcurrency 是 worker 并发数。
	TaskTimeout       time.Duration // TaskTimeout 是单个任务超时时间。
	ShutdownTimeout   time.Duration // ShutdownTimeout 是关闭等待时间。
}

// NewConfig 将 protobuf 配置转换为任务运行时配置。
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
