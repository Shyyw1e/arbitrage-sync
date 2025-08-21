package redisqueue

import (
	"context"
	"time"

	"github.com/Shyyw1e/arbitrage-sync/pkg/logger"
	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func InitRedisClient() error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "redis-internal:6379",
		Password: "",
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if pong, err := RedisClient.Ping(ctx).Result(); err != nil {
		logger.Log.Errorf("failed to ping redis: %v", err)
		return err
	} else {
		logger.Log.Infof("Redis connected: %s", pong)
	}
	return nil
}
