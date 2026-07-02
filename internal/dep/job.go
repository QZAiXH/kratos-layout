package dep

import (
	"github.com/go-redsync/redsync/v4"
	goredis "github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

// NewAsynqRedisConnOpt 从 Redis 客户端生成 Asynq 连接配置。
func NewAsynqRedisConnOpt(client *redis.Client) asynq.RedisConnOpt {
	if client == nil {
		return nil
	}
	return asynq.RedisClientOpt{
		Addr:     client.Options().Addr,
		Username: client.Options().Username,
		Password: client.Options().Password,
		DB:       client.Options().DB,
	}
}

// NewAsynqClient 从 Redis 客户端创建 Asynq 入队客户端。
func NewAsynqClient(client *redis.Client) *asynq.Client {
	if client == nil {
		return nil
	}
	return asynq.NewClientFromRedisClient(client)
}

// NewRedsync 从 Redis 客户端创建分布式锁管理器。
func NewRedsync(client *redis.Client) *redsync.Redsync {
	if client == nil {
		return nil
	}
	return redsync.New(goredis.NewPool(client))
}
