package redisqueue

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Shyyw1e/arbitrage-sync/internal/core/domain"
	"github.com/Shyyw1e/arbitrage-sync/internal/infrastructure/db"
	"github.com/Shyyw1e/arbitrage-sync/pkg/logger"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const JobQueueKey = "jobs:queue"

func EnqueueJob(job string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RedisClient.RPush(ctx, JobQueueKey, job).Err(); err != nil {
		logger.Log.Errorf("failed to enqueue job: %v", err)
		return err
	}
	logger.Log.Infof("Enqueued job: %s", job)
	return nil
}

func StartAnalysisForUser(bot *tgbotapi.BotAPI, chatID int64, userState *domain.UserState) error {
	userState.Step = "ready_to_run"
	if err := userStore.Set(chatID, userState); err != nil {
		logger.Log.Errorf("failed to set user state: %v", err)
		return err
	}
	job := fmt.Sprintf("detect-as:%.2f:%.2f:%d", userState.MinDiff, userState.MaxSum, chatID)
	if err := EnqueueJob(job); err != nil {
		logger.Log.Errorf("failed to enqueue job: %v", err)
		return err
	}
	return nil
}

func StopAnalysis(store db.UserStatesStore, chatID int64) error {
	running := dispatcher.isRunning(chatID)

	queued, _ := hasJobsForChat(chatID)

	st, _ := store.Get(chatID)
	step := ""
	if st != nil {
		step = st.Step
	}

	if !running && !queued && step != "ready_to_run" {
		logger.Log.Infof("Stop requested but nothing to stop chatID=%d", chatID)
		return fmt.Errorf("no running analysis for chatID %d", chatID)
	}

	if st != nil {
		st.Step = "not_active"
		_ = store.Set(chatID, st)
	}

	if err := dispatcher.stop(chatID, store); err != nil {
		logger.Log.Errorf("dispatcher.stop failed: %v", err)
	}

	_ = removeJobsForChat(chatID)

	logger.Log.Infof("Analysis stopped for chatID %d (running=%v queued=%v prevStep=%s)", chatID, running, queued, step)
	return nil
}


func hasJobsForChat(chatID int64) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	vals, err := RedisClient.LRange(ctx, JobQueueKey, 0, -1).Result()
	if err != nil {
		return false, err
	}
	suffix := ":" + strconv.FormatInt(chatID, 10)
	for _, v := range vals {
		if strings.HasSuffix(v, suffix) {
			return true, nil
		}
	}
	return false, nil
}

func removeJobsForChat(chatID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	vals, err := RedisClient.LRange(ctx, JobQueueKey, 0, -1).Result()
	if err != nil {
		return err
	}
	suffix := ":" + strconv.FormatInt(chatID, 10)
	for _, v := range vals {
		if strings.HasSuffix(v, suffix) {
			_ = RedisClient.LRem(ctx, JobQueueKey, 0, v).Err()
		}
	}
	return nil
}
