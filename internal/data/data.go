package data

import (
	"github.com/QZAiXH/kratos-layout/internal/data/base"

	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(base.NewData, NewTodoRepo)

type Data = base.Data
