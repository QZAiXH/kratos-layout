package biz

import (
	todobiz "github.com/QZAiXH/kratos-layout/internal/biz/todo"

	"github.com/google/wire"
)

// ProviderSet 提供 biz 层依赖注入集合。
var ProviderSet = wire.NewSet(todobiz.NewUseCase)
