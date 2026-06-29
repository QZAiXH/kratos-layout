package dep

import (
	"testing"

	"helloworld/internal/pkg/authz"
)

// TestNewCasbinEnforcerAllowsAdmin 验证 Casbin 初始化后内置管理员通配策略。
func TestNewCasbinEnforcerAllowsAdmin(t *testing.T) {
	enforcer, err := NewCasbinEnforcer()
	if err != nil {
		t.Fatalf("NewCasbinEnforcer() error = %v", err)
	}

	allowed, err := enforcer.Enforce(authz.RoleAdmin, "/anything", authz.ActionInvoke)
	if err != nil {
		t.Fatalf("Enforce(admin) error = %v", err)
	}
	if !allowed {
		t.Fatal("Enforce(admin) = false, want true")
	}
}
