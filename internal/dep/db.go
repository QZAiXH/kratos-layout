package dep

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/QZAiXH/kratos-layout/internal/conf"
	"github.com/QZAiXH/kratos-layout/internal/data/ent"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func NewDB(c *conf.Data, logger *slog.Logger) (*ent.Client, func(), error) {
	dbConf := c.GetDatabase()
	if dbConf == nil {
		return nil, func() {}, nil
	}
	driver := strings.TrimSpace(dbConf.GetDriver())
	source := strings.TrimSpace(dbConf.GetSource())
	if driver == "" || source == "" {
		return nil, func() {}, nil
	}

	client, err := ent.Open(driver, source)
	if err != nil {
		return nil, nil, fmt.Errorf("open ent database: %w", err)
	}
	cleanup := func() {
		if err := client.Close(); err != nil && logger != nil {
			logger.Warn("close ent database", slog.Any("error", err))
		}
	}
	if dbConf.GetAutoMigrate() {
		if err := client.Schema.Create(context.Background()); err != nil {
			cleanup()
			return nil, nil, fmt.Errorf("auto migrate ent schema: %w", err)
		}
	}
	return client, cleanup, nil
}

func NewRedis(c *conf.Data, logger *slog.Logger) (*redis.Client, func(), error) {
	redisConf := c.GetRedis()
	if redisConf == nil || strings.TrimSpace(redisConf.GetAddr()) == "" {
		return nil, func() {}, nil
	}

	options := &redis.Options{
		Network:  strings.TrimSpace(redisConf.GetNetwork()),
		Addr:     strings.TrimSpace(redisConf.GetAddr()),
		Password: redisConf.GetPassword(),
		DB:       int(redisConf.GetDb()),
	}
	if timeout := redisConf.GetDialTimeout(); timeout != nil {
		options.DialTimeout = timeout.AsDuration()
	}
	if timeout := redisConf.GetReadTimeout(); timeout != nil {
		options.ReadTimeout = timeout.AsDuration()
	}
	if timeout := redisConf.GetWriteTimeout(); timeout != nil {
		options.WriteTimeout = timeout.AsDuration()
	}

	client := redis.NewClient(options)
	return client, func() {
		if err := client.Close(); err != nil && logger != nil {
			logger.Warn("close redis client", slog.Any("error", err))
		}
	}, nil
}
