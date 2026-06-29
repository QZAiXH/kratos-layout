package base

import (
	"context"
	"log/slog"
)

// Transaction runs a business operation in a transaction.
type Transaction interface {
	InTx(context.Context, func(ctx context.Context) error) error
}

// UseCase holds shared biz-layer dependencies.
type UseCase struct {
	Tx  Transaction
	Log *slog.Logger
}

// NewUseCase creates shared biz-layer dependencies.
func NewUseCase(tx Transaction, logger *slog.Logger, moduleName string) *UseCase {
	if logger == nil {
		logger = slog.Default()
	}
	if moduleName != "" {
		logger = logger.With("module", moduleName)
	}
	return &UseCase{Tx: tx, Log: logger}
}
