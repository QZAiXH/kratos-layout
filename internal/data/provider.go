package data

import (
	"github.com/QZAiXH/kratos-layout/internal/data/base"

	"github.com/google/wire"
)

// ProviderSet 提供 data 层依赖注入集合。
var ProviderSet = wire.NewSet(base.NewData, NewTodoRepo)

type Data = base.Data
