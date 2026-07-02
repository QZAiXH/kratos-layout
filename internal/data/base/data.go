package base

import (
	"github.com/QZAiXH/kratos-layout/internal/data/ent"

	"github.com/casbin/casbin/v3"
	"github.com/redis/go-redis/v9"
)

// Data 聚合数据层共享依赖。
type Data struct {
	DB   *ent.Database    // DB 是 Ent 数据库包装客户端。
	RDB  *redis.Client    // RDB 是 Redis 客户端。
	Rule *casbin.Enforcer // Rule 是 Casbin 鉴权执行器。
}

// NewData 创建数据层共享依赖容器。
func NewData(db *ent.Database, redisClient *redis.Client, rule *casbin.Enforcer) (*Data, error) {
	return &Data{DB: db, RDB: redisClient, Rule: rule}, nil
}
