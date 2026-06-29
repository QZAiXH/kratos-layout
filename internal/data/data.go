package data

import (
	"database/sql"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewTodoRepo)

// Data owns shared data-layer clients.
type Data struct {
	DB  *sql.DB
	RDB *redis.Client
}

// NewData aggregates initialized data-layer clients.
func NewData(db *sql.DB, rdb *redis.Client) (*Data, error) {
	return &Data{DB: db, RDB: rdb}, nil
}
