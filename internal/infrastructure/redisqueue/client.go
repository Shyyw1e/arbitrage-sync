package redisqueue

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/Shyyw1e/arbitrage-sync/pkg/logger"
	"github.com/redis/go-redis/v9"
)

var (
    redisOpts *redis.Options
    redisMu   sync.RWMutex
    RedisClient *redis.Client
)

func InitRedisClient() error {
    redisOpts = &redis.Options{
        Addr: "redis-internal:6379",
        Password: "",
        DB: 0,
    }
    return resetRedisClient()
}

func resetRedisClient() error {
    c := redis.NewClient(redisOpts)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := c.Ping(ctx).Err(); err != nil {
        return err
    }
    redisMu.Lock()
    if RedisClient != nil { _ = RedisClient.Close() }
    RedisClient = c
    redisMu.Unlock()
    logger.Log.Info("Redis client reinitialized")
    return nil
}

func getRedis() *redis.Client {
    redisMu.RLock()
    c := RedisClient
    redisMu.RUnlock()
    return c
}

func isReadOnlyErr(err error) bool {
    if err == nil { return false }
    s := strings.ToLower(err.Error())
    return strings.Contains(s, "readonly")
}