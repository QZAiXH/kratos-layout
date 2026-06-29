package dep

import "github.com/google/wire"

// ProviderSet is external infrastructure providers.
var ProviderSet = wire.NewSet(NewDB, NewRedis, NewCasbinEnforcer)
