package base

import "log/slog"

// UseCase 提供业务用例通用依赖。
type UseCase struct {
	Log *slog.Logger // Log 是带模块字段的业务日志器。
}

// NewUseCase 创建带模块日志字段的业务用例基类。
func NewUseCase(logger *slog.Logger, module string) *UseCase {
	if logger != nil {
		logger = logger.With(slog.String("module", module))
	}
	return &UseCase{Log: logger}
}
