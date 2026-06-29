package dep

import (
	"database/sql"
	"fmt"
	"strings"

	"helloworld/internal/conf"

	"github.com/go-kratos/kratos/v3/log"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func NewDB(c *conf.Data) (*sql.DB, func(), error) {
	dbConf := c.GetDatabase()
	if dbConf == nil {
		return nil, func() {}, nil
	}
	driver := strings.TrimSpace(dbConf.GetDriver())
	source := strings.TrimSpace(dbConf.GetSource())
	if driver == "" || source == "" {
		return nil, func() {}, nil
	}

	db, err := sql.Open(driver, source)
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}
	return db, func() {
		if err := db.Close(); err != nil {
			log.Warnf("close database client: %v", err)
		}
	}, nil
}

func NewRedis(c *conf.Data) (*redis.Client, func(), error) {
	redisConf := c.GetRedis()
	if redisConf == nil {
		return nil, func() {}, nil
	}
	addr := strings.TrimSpace(redisConf.GetAddr())
	if addr == "" {
		return nil, func() {}, nil
	}

	options := &redis.Options{
		Addr:     addr,
		Password: redisConf.GetPassword(),
	}
	if network := strings.TrimSpace(redisConf.GetNetwork()); network != "" {
		options.Network = network
	}
	if timeout := redisConf.GetReadTimeout(); timeout != nil {
		options.ReadTimeout = timeout.AsDuration()
	}
	if timeout := redisConf.GetWriteTimeout(); timeout != nil {
		options.WriteTimeout = timeout.AsDuration()
	}

	client := redis.NewClient(options)
	return client, func() {
		if err := client.Close(); err != nil {
			log.Warnf("close redis client: %v", err)
		}
	}, nil
}
