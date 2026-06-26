package server

import (
	"helloworld/internal/pkg/authz"

	"github.com/google/wire"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(authz.NewCasbinEnforcer, NewTokenManager, NewSecurityMiddleware, NewGRPCServer, NewHTTPServer)
