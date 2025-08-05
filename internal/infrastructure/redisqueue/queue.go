package redisqueue

import (
	"context"
	"fmt"
	"time"

	"github.com/Shyyw1e/arbitrage-sync/internal/core/domain"
	"github.com/Shyyw1e/arbitrage-sync/pkg/logger"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/redis/go-redis/v9"
)

const JobQueueKey = "jobs:queue"

func EnqueueJob(job string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := RedisClient.RPush(ctx, JobQueueKey, job).Err()
	if err != nil {
		logger.Log.Errorf("failed to enqueue job: %v", err)
		return err
	}
	logger.Log.Infof("Enqueued job: %v", job)
	return nil
}

func DequeueJob() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	job, err := RedisClient.LPop(ctx, JobQueueKey).Result()
	if err == redis.Nil {
		logger.Log.Info("empty queue")
		return "", nil
	}
	if err != nil {
		logger.Log.Errorf("failed to dequeue job: %v", err)
		return "", err
	}

	logger.Log.Infof("Dequeued job: %s", job)
	return job, nil
}

func QueueLength() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	length, err := RedisClient.LLen(ctx, JobQueueKey).Result()
	if err != nil {
		logger.Log.Errorf("failed to get queue length: %v", err)
		return 0, err
	}

	logger.Log.Infof("Got queue length: %v", length)
	return length, nil
}

func StartAnalysisForUser(bot *tgbotapi.BotAPI, chatID int64, userState *domain.UserState) error {
	job := fmt.Sprintf("detect-as:%.2f:%.2f:%d", userState.MinDiff, userState.MaxSum, chatID)
	err := EnqueueJob(job)
	if err != nil {
		logger.Log.Errorf("failed to enqueue job: %v", err)
		return err
	}

	return nil
}