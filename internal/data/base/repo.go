package base

import "log/slog"

// Repo holds shared repo dependencies.
type Repo struct {
	Data *Data
	Log  *slog.Logger
}

// NewRepo creates shared repo dependencies.
func NewRepo(data *Data, logger *slog.Logger, moduleName string) *Repo {
	if logger == nil {
		logger = slog.Default()
	}
	if moduleName != "" {
		logger = logger.With("module", moduleName)
	}
	return &Repo{Data: data, Log: logger}
}
