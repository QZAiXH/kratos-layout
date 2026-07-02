package service

import "github.com/google/wire"

// ProviderSet 提供 service 层依赖注入集合。
var ProviderSet = wire.NewSet(NewTodoService)
