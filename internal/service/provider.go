package service

import (
	todoservice "github.com/QZAiXH/kratos-layout/internal/service/todo"

	"github.com/google/wire"
)

// ProviderSet 提供 service 层依赖注入集合。
var ProviderSet = wire.NewSet(todoservice.NewService)
