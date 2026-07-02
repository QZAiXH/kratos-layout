package base

import "log/slog"

type UseCase struct {
	Log *slog.Logger
}

func NewUseCase(logger *slog.Logger, module string) *UseCase {
	if logger != nil {
		logger = logger.With(slog.String("module", module))
	}
	return &UseCase{Log: logger}
}
