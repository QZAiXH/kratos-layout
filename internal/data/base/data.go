package base

import (
	"github.com/QZAiXH/kratos-layout/internal/data/ent"

	"github.com/casbin/casbin/v3"
	"github.com/redis/go-redis/v9"
)

type Data struct {
	DB   *ent.Client
	RDB  *redis.Client
	Rule *casbin.Enforcer
}

func NewData(db *ent.Client, redisClient *redis.Client, rule *casbin.Enforcer) (*Data, error) {
	return &Data{DB: db, RDB: redisClient, Rule: rule}, nil
}
