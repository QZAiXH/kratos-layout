package base

import "log/slog"

// Repo 提供仓储实现通用依赖。
type Repo struct {
	Data *Data        // Data 是数据层共享依赖。
	Log  *slog.Logger // Log 是带模块字段的仓储日志器。
}

// NewRepo 创建带模块日志字段的仓储基类。
func NewRepo(data *Data, logger *slog.Logger, module string) *Repo {
	if logger != nil {
		logger = logger.With(slog.String("module", module))
	}
	return &Repo{Data: data, Log: logger}
}
