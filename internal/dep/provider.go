package dep

import "github.com/google/wire"

var ProviderSet = wire.NewSet(NewDB, NewRedis, NewCasbinEnforcer, NewAsynqRedisConnOpt, NewAsynqClient, NewRedsync)
