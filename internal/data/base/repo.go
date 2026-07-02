package base

import "log/slog"

type Repo struct {
	Data *Data
	Log  *slog.Logger
}

func NewRepo(data *Data, logger *slog.Logger, module string) *Repo {
	if logger != nil {
		logger = logger.With(slog.String("module", module))
	}
	return &Repo{Data: data, Log: logger}
}
