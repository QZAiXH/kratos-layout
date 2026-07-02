package dep

import (
	"github.com/QZAiXH/kratos-layout/internal/pkg/authz"

	casbinv3 "github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
)

// NewCasbinEnforcer 创建默认 Casbin 执行器。
func NewCasbinEnforcer() (*casbinv3.Enforcer, error) {
	m, err := model.NewModelFromString(casbinModel)
	if err != nil {
		return nil, err
	}
	enforcer, err := casbinv3.NewEnforcer(m)
	if err != nil {
		return nil, err
	}
	_, err = enforcer.AddPolicy(authz.RoleAdmin, "*", "*")
	return enforcer, err
}

const casbinModel = `
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
