package data

import (
	"github.com/QZAiXH/kratos-layout/internal/data/base"
	tododata "github.com/QZAiXH/kratos-layout/internal/data/todo"

	"github.com/google/wire"
)

// ProviderSet 提供 data 层依赖注入集合。
var ProviderSet = wire.NewSet(base.NewData, tododata.NewRepo)

type Data = base.Data
