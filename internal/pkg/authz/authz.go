package authz

import (
	"context"
	"strings"

	"github.com/QZAiXH/kratos-layout/internal/pkg/token"

	casbinv3 "github.com/casbin/casbin/v3"
	kratoserrors "github.com/go-kratos/kratos/v3/errors"
	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/transport"
)

const (
	RoleAdmin    = "admin"
	RoleUser     = "user"
	ActionInvoke = "invoke"
)

const (
	authorizationKey = "Authorization"
	bearerWord       = "Bearer"
)

var (
	ErrUnauthorized = kratoserrors.Unauthorized("UNAUTHORIZED", "unauthorized")
	ErrForbidden    = kratoserrors.Forbidden("FORBIDDEN", "forbidden")
)

// claimsContextKey 是访问令牌声明的 context key。
type claimsContextKey struct{}

// SecurityUser 表示 Casbin 鉴权请求主体。
type SecurityUser struct {
	Subject string // Subject 是用户或角色主体。
	Object  string // Object 是受保护的操作对象。
	Action  string // Action 是访问动作。
}

// ContextWithAccessTokenClaims 将访问令牌声明写入 context。
func ContextWithAccessTokenClaims(ctx context.Context, claims *token.AccessTokenClaims) context.Context {
	if claims == nil {
		return ctx
	}
	return context.WithValue(ctx, claimsContextKey{}, claims)
}

// GetAccessTokenClaims 从 context 读取访问令牌声明。
func GetAccessTokenClaims(ctx context.Context) (*token.AccessTokenClaims, bool) {
	claims, ok := ctx.Value(claimsContextKey{}).(*token.AccessTokenClaims)
	return claims, ok
}

// GetUserIDFromContext 从 context 中读取当前用户 ID。
func GetUserIDFromContext(ctx context.Context) (string, error) {
	claims, ok := GetAccessTokenClaims(ctx)
	if !ok || strings.TrimSpace(claims.UserID) == "" {
		return "", ErrUnauthorized
	}
	return claims.UserID, nil
}

// JWTServer 创建 JWT 服务端认证中间件。
func JWTServer(manager *token.Manager) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			if manager == nil {
				return nil, ErrUnauthorized
			}
			accessToken, ok := bearerTokenFromContext(ctx)
			if !ok {
				return nil, ErrUnauthorized
			}
			claims, err := manager.ValidateAccessToken(accessToken)
			if err != nil {
				return nil, err
			}
			return handler(ContextWithAccessTokenClaims(ctx, claims), req)
		}
	}
}

// CasbinServer 创建 Casbin 服务端鉴权中间件。
func CasbinServer(enforcer *casbinv3.Enforcer) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			if enforcer == nil {
				return nil, ErrForbidden
			}
			user, err := NewSecurityUser(ctx)
			if err != nil {
				return nil, err
			}
			allowed, err := enforcer.Enforce(user.Subject, user.Object, user.Action)
			if err != nil {
				return nil, err
			}
			if !allowed {
				return nil, ErrForbidden
			}
			return handler(ctx, req)
		}
	}
}

// NewSecurityUser 根据当前请求上下文生成 Casbin 鉴权主体。
func NewSecurityUser(ctx context.Context) (*SecurityUser, error) {
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	tr, ok := transport.FromServerContext(ctx)
	if !ok || strings.TrimSpace(tr.Operation()) == "" {
		return nil, ErrForbidden
	}
	return &SecurityUser{Subject: userID, Object: tr.Operation(), Action: ActionInvoke}, nil
}

// bearerTokenFromContext 从请求头提取 Bearer token。
func bearerTokenFromContext(ctx context.Context) (string, bool) {
	tr, ok := transport.FromServerContext(ctx)
	if !ok || tr.RequestHeader() == nil {
		return "", false
	}
	auth := strings.TrimSpace(tr.RequestHeader().Get(authorizationKey))
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], bearerWord) || strings.TrimSpace(parts[1]) == "" {
		return "", false
	}
	return strings.TrimSpace(parts[1]), true
}
