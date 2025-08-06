package redisqueue

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Shyyw1e/arbitrage-sync/internal/core/usecase"
	"github.com/Shyyw1e/arbitrage-sync/internal/infrastructure/db"
	"github.com/Shyyw1e/arbitrage-sync/pkg/logger"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	userStore         db.UserStatesStore
	runningTasks      = make(map[int64]context.CancelFunc)
	runningTasksMutex sync.Mutex
)

func InitRedisQueue(store db.UserStatesStore) {
	userStore = store
}

func StartWorkerLoop(bot *tgbotapi.BotAPI) {
	
	go func ()  {
		for {
			job, err := DequeueJob()
			if err != nil {
				logger.Log.Errorf("failed to dequeue job: %v", err)
				time.Sleep(2 * time.Second)
				continue
			}

			if job == "" {
				logger.Log.Info("Empty job")
				time.Sleep(30 * time.Second)
				continue
			}

			handleJob(job, bot)
		}
	}()
}

func handleJob(job string, bot *tgbotapi.BotAPI) {
	if !strings.HasPrefix(job, "detect-as:") {
		logger.Log.Errorf("unknown job format: %v", job)
		return
	}

	parts := strings.Split(job, ":")
	if len(parts) != 4 {
		logger.Log.Errorf("invalid job parts: %v", job)
		return
	}

	minDiff, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		logger.Log.Errorf("failed to convert minDiff: %v", err)
		return
	}
	minDiff = math.Round(minDiff*100) / 100

	maxSum, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		logger.Log.Errorf("failed to convert maxSum: %v", err)
		return
	}
	maxSum = math.Round(maxSum*100) / 100

	chatID, err := strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		logger.Log.Errorf("failed to convert chatID: %v", err)
		return
	}

	runningTasksMutex.Lock()
	if _, exists := runningTasks[chatID]; exists {
		runningTasksMutex.Unlock()
		logger.Log.Infof("Analysis already running for chatID %d", chatID)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	runningTasks[chatID] = cancel
	runningTasksMutex.Unlock()

	go func() {
		defer func() {
			runningTasksMutex.Lock()
			delete(runningTasks, chatID)
			runningTasksMutex.Unlock()
			logger.Log.Infof("Stopped analysis loop for chatID %d", chatID)
		}()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				state, err := userStore.Get(chatID)
				logger.Log.Infof("Current state for user %d: %+v", chatID, state)
				if err != nil || state == nil || state.Step != "ready_to_run" {
					logger.Log.Infof("User %d is not active. Stopping analysis.", chatID)
					return
				}

				ops, pots, err := usecase.DetectAS(minDiff, maxSum)
				if err != nil {
					logger.Log.Errorf("failed to execute DetectAS: %v", err)
					time.Sleep(15 * time.Second)
					continue
				}

				if len(ops) == 0 && len(pots) == 0 {
					bot.Send(tgbotapi.NewMessage(chatID, "🤔 Арбитражных ситуаций не найдено."))
				} else {
					for _, op := range ops {
						text := fmt.Sprintf("💰 Найден арбитраж!\nBuy %s @ %.2f\nSell %s @ %.2f\nProfit: %.2f",
							op.BuyExchange, op.BuyPrice, op.SellExchange, op.SellPrice, op.ProfitMargin)
						bot.Send(tgbotapi.NewMessage(chatID, text))
					}
					for _, op := range pots {
						text := fmt.Sprintf("💰 Найден обратный арбитраж!\nBuy %s @ %.2f\nSell %s @ %.2f\nProfit: %.2f",
							op.BuyExchange, op.BuyPrice, op.SellExchange, op.SellPrice, op.ProfitMargin)
						bot.Send(tgbotapi.NewMessage(chatID, text))
					}
				}

				time.Sleep(5 * time.Second)
			}
		}
	}()
}

func StopAnalysis(store db.UserStatesStore, chatID int64) error {
	runningTasksMutex.Lock()
	cancel, ok := runningTasks[chatID]
	if ok {
		cancel()
		delete(runningTasks, chatID)
	}
	runningTasksMutex.Unlock()

	if !ok {
		logger.Log.Infof("No running analysis for chatID %d", chatID)
		return fmt.Errorf("no running analysis for chatID %d", chatID)
	}

	state, err := store.Get(chatID)
	if err == nil && state != nil {
		state.Step = "not_active"
		store.Set(chatID, state)
	}

	logger.Log.Infof("Analysis stopped for chatID %d", chatID)
	return nil
}