package base

import (
	"database/sql"

	"github.com/casbin/casbin/v3"
	"github.com/redis/go-redis/v9"
)

// Data aggregates shared data-layer clients.
type Data struct {
	DB   *sql.DB
	RDB  *redis.Client
	Rule *casbin.Enforcer
}

// NewData aggregates initialized data-layer clients.
func NewData(db *sql.DB, rdb *redis.Client, rule *casbin.Enforcer) (*Data, error) {
	return &Data{DB: db, RDB: rdb, Rule: rule}, nil
}
