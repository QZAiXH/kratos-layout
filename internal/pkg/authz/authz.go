package authz

import (
	"context"
	"strings"

	"helloworld/internal/pkg/token"

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

type claimsContextKey struct{}

type SecurityUser struct {
	Subject string
	Object  string
	Action  string
}

func ContextWithAccessTokenClaims(ctx context.Context, claims *token.AccessTokenClaims) context.Context {
	if claims == nil {
		return ctx
	}
	return context.WithValue(ctx, claimsContextKey{}, claims)
}

func GetAccessTokenClaims(ctx context.Context) (*token.AccessTokenClaims, bool) {
	claims, ok := ctx.Value(claimsContextKey{}).(*token.AccessTokenClaims)
	return claims, ok
}

func GetUserIDFromContext(ctx context.Context) (string, error) {
	claims, ok := GetAccessTokenClaims(ctx)
	if !ok || strings.TrimSpace(claims.UserID) == "" {
		return "", ErrUnauthorized
	}
	return claims.UserID, nil
}

func GetOptionalUserIDFromContext(ctx context.Context) string {
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		return ""
	}
	return userID
}

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

func NewSecurityUser(ctx context.Context) (*SecurityUser, error) {
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	tr, ok := transport.FromServerContext(ctx)
	if !ok || strings.TrimSpace(tr.Operation()) == "" {
		return nil, ErrForbidden
	}
	return &SecurityUser{
		Subject: userID,
		Object:  tr.Operation(),
		Action:  ActionInvoke,
	}, nil
}

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
