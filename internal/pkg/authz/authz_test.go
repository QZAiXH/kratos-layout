package authz

import (
	"context"
	"testing"
	"time"

	"helloworld/internal/pkg/token"

	casbinv3 "github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/transport"
)

// TestJWTAndCasbinMiddleware 验证 JWT 认证和 Casbin 授权通过后会注入当前用户。
func TestJWTAndCasbinMiddleware(t *testing.T) {
	manager, err := token.NewManager("", time.Minute, time.Hour)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	pair, err := manager.GenerateTokenPair("user_1", "v1")
	if err != nil {
		t.Fatalf("GenerateTokenPair() error = %v", err)
	}
	enforcer, err := newTestCasbinEnforcer()
	if err != nil {
		t.Fatalf("NewCasbinEnforcer() error = %v", err)
	}
	if _, err := enforcer.AddGroupingPolicy("user_1", RoleAdmin); err != nil {
		t.Fatalf("AddGroupingPolicy() error = %v", err)
	}

	handler := middleware.Chain(JWTServer(manager), CasbinServer(enforcer))(func(ctx context.Context, _ any) (any, error) {
		return GetUserIDFromContext(ctx)
	})
	ctx := transport.NewServerContext(context.Background(), fakeTransport{
		operation: "/todo.v1.TodoService/CreateTodo",
		header:    fakeHeader{"Authorization": "Bearer " + pair.AccessToken},
	})
	reply, err := handler(ctx, nil)
	if err != nil {
		t.Fatalf("middleware() error = %v", err)
	}
	if reply != "user_1" {
		t.Fatalf("middleware() reply = %v, want user_1", reply)
	}
}

// TestJWTAndCasbinMiddlewareRejectsMissingToken 验证缺少令牌时中间件会拒绝请求。
func TestJWTAndCasbinMiddlewareRejectsMissingToken(t *testing.T) {
	manager, err := token.NewManager("", time.Minute, time.Hour)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	enforcer, err := newTestCasbinEnforcer()
	if err != nil {
		t.Fatalf("NewCasbinEnforcer() error = %v", err)
	}

	handler := middleware.Chain(JWTServer(manager), CasbinServer(enforcer))(func(context.Context, any) (any, error) {
		return nil, nil
	})
	ctx := transport.NewServerContext(context.Background(), fakeTransport{
		operation: "/protected.Service/Call",
		header:    fakeHeader{},
	})
	if _, err := handler(ctx, nil); err == nil {
		t.Fatal("middleware() error = nil, want error")
	}
}

type fakeTransport struct {
	operation string
	header    fakeHeader
}

func newTestCasbinEnforcer() (*casbinv3.Enforcer, error) {
	m, err := model.NewModelFromString(testCasbinModel)
	if err != nil {
		return nil, err
	}
	enforcer, err := casbinv3.NewEnforcer(m)
	if err != nil {
		return nil, err
	}
	_, err = enforcer.AddPolicy(RoleAdmin, "*", "*")
	return enforcer, err
}

func (t fakeTransport) Kind() transport.Kind            { return transport.KindHTTP }
func (t fakeTransport) Endpoint() string                { return "" }
func (t fakeTransport) Operation() string               { return t.operation }
func (t fakeTransport) RequestHeader() transport.Header { return t.header }
func (t fakeTransport) ReplyHeader() transport.Header   { return fakeHeader{} }

type fakeHeader map[string]string

func (h fakeHeader) Get(key string) string      { return h[key] }
func (h fakeHeader) Set(key, value string)      { h[key] = value }
func (h fakeHeader) Add(key, value string)      { h[key] = value }
func (h fakeHeader) Keys() []string             { return nil }
func (h fakeHeader) Values(key string) []string { return []string{h[key]} }

const testCasbinModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = (r.sub == p.sub || g(r.sub, p.sub)) && (p.obj == "*" || r.obj == p.obj) && (p.act == "*" || r.act == p.act)
`
