package server

import (
	"context"
	"time"

	v1 "github.com/QZAiXH/kratos-layout/api/todo/v1"
	"github.com/QZAiXH/kratos-layout/internal/conf"
	"github.com/QZAiXH/kratos-layout/internal/pkg/authz"
	"github.com/QZAiXH/kratos-layout/internal/pkg/token"

	casbinv3 "github.com/casbin/casbin/v3"
	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/middleware/selector"
)

func NewTokenManager(c *conf.Auth) (*token.Manager, error) {
	var privateKeyPath string
	var accessTTL, refreshTTL time.Duration
	if c != nil {
		privateKeyPath = c.PrivateKeyPath
		if c.AccessTokenTtl != nil {
			accessTTL = c.AccessTokenTtl.AsDuration()
		}
		if c.RefreshTokenTtl != nil {
			refreshTTL = c.RefreshTokenTtl.AsDuration()
		}
	}
	return token.NewManager(privateKeyPath, accessTTL, refreshTTL)
}

func NewSecurityMiddleware(manager *token.Manager, enforcer *casbinv3.Enforcer) middleware.Middleware {
	return selector.Server(
		authz.JWTServer(manager),
		authz.CasbinServer(enforcer),
	).Match(NewProtectedMatcher()).Build()
}

func NewProtectedMatcher() selector.MatchFunc {
	anonymous := map[string]struct{}{
		v1.OperationTodoServiceCreateTodo: {},
		v1.OperationTodoServiceDeleteTodo: {},
		v1.OperationTodoServiceGetTodo:    {},
		v1.OperationTodoServiceListTodos:  {},
		v1.OperationTodoServiceSyncTodos:  {},
		v1.OperationTodoServiceUpdateTodo: {},
		v1.OperationTodoServiceWatchTodos: {},
	}
	return func(_ context.Context, operation string) bool {
		_, ok := anonymous[operation]
		return !ok
	}
}
