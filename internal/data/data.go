package data

import (
	"database/sql"
	"fmt"
	"strings"

	"helloworld/internal/conf"

	"github.com/go-kratos/kratos/v3/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/wire"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewTodoRepo)

// Data owns shared data-layer clients.
type Data struct {
	DB  *sql.DB
	RDB *redis.Client
}

// NewData creates shared data-layer clients from runtime config.
func NewData(c *conf.Data) (*Data, func(), error) {
	db, err := openDatabase(c.GetDatabase())
	if err != nil {
		return nil, nil, err
	}
	rdb := newRedisClient(c.GetRedis())

	data := &Data{
		DB:  db,
		RDB: rdb,
	}
	cleanup := func() {
		log.Info("closing the data resources")
		if data.RDB != nil {
			if err := data.RDB.Close(); err != nil {
				log.Warnf("close redis client: %v", err)
			}
		}
		if data.DB != nil {
			if err := data.DB.Close(); err != nil {
				log.Warnf("close database client: %v", err)
			}
		}
	}
	return data, cleanup, nil
}

func openDatabase(c *conf.Data_Database) (*sql.DB, error) {
	if c == nil {
		return nil, nil
	}
	driver := strings.TrimSpace(c.GetDriver())
	source := strings.TrimSpace(c.GetSource())
	if driver == "" || source == "" {
		return nil, nil
	}

	db, err := sql.Open(driver, source)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	return db, nil
}

func newRedisClient(c *conf.Data_Redis) *redis.Client {
	if c == nil {
		return nil
	}
	addr := strings.TrimSpace(c.GetAddr())
	if addr == "" {
		return nil
	}

	options := &redis.Options{
		Addr:     addr,
		Password: c.GetPassword(),
	}
	if network := strings.TrimSpace(c.GetNetwork()); network != "" {
		options.Network = network
	}
	if timeout := c.GetReadTimeout(); timeout != nil {
		options.ReadTimeout = timeout.AsDuration()
	}
	if timeout := c.GetWriteTimeout(); timeout != nil {
		options.WriteTimeout = timeout.AsDuration()
	}
	return redis.NewClient(options)
}
