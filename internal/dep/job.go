package dep

import (
	"github.com/go-redsync/redsync/v4"
	goredis "github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

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

func NewAsynqClient(client *redis.Client) *asynq.Client {
	if client == nil {
		return nil
	}
	return asynq.NewClientFromRedisClient(client)
}

func NewRedsync(client *redis.Client) *redsync.Redsync {
	if client == nil {
		return nil
	}
	return redsync.New(goredis.NewPool(client))
}
