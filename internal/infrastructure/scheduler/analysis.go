package scheduler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Shyyw1e/arbitrage-sync/internal/core/domain"
	"github.com/Shyyw1e/arbitrage-sync/internal/core/usecase"
	"github.com/Shyyw1e/arbitrage-sync/pkg/logger"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var RunningTasks = make(map[int64]context.CancelFunc)
var tasksMU sync.Mutex

func StartAnalysisForUser(bot *tgbotapi.BotAPI, chatID int64, userState *domain.UserState) error {
	tasksMU.Lock()
	if _, ok := RunningTasks[chatID]; ok {
		logger.Log.Error("task is running")
		if _, err := bot.Send(tgbotapi.NewMessage(chatID, "Анализ уже запущен")); err != nil {
			logger.Log.Errorf("failed to send message: %v", err)
			return err
		}
		return nil
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	RunningTasks[chatID] = cancel
	tasksMU.Unlock()

	go func () {
		for {
			select {
			case <-ctx.Done():
				tasksMU.Lock()
				delete(RunningTasks, chatID)
				tasksMU.Unlock()

				logger.Log.Info("analysis finished, context cancelled")
				if _, err := bot.Send(tgbotapi.NewMessage(chatID, "Анализ завершен")); err != nil {
					logger.Log.Errorf("failed to send message: %v", err)
					return
				}
				return

			default:
				opps, pots, err := usecase.DetectAS(userState.MinDiff, userState.MaxSum, chatID)
				if err != nil {
					logger.Log.Errorf("failed to detect: %v", err)
					continue 
				}
				for _, opp := range opps {
					msgAS := fmt.Sprintf("Арбитражная ситуация:\nExchange зеленого стакана: %v\nExchange красного стакана: %v\nPrice зеленого стакана: %v\nPrice красного стакана: %v\nAmount: %v\nProfit Margin: %v\nExpecting income: %v",
						 opp.BuyExchange, opp.SellExchange, opp.BuyPrice, opp.SellPrice, opp.BuyAmount, opp.ProfitMargin, opp.BuyAmount * (1 + opp.ProfitMargin))
					if _, err := bot.Send(tgbotapi.NewMessage(chatID, msgAS)); err != nil {
						logger.Log.Errorf("failed to send the message:%v", err)
						continue 
					}
				}

				for _, pot := range pots {
					msgAS := fmt.Sprintf("Потенциальная ситуация:\nExchange зеленого стакана: %v\nExchange красного стакана: %v\nPrice зеленого стакана: %v\nPrice красного стакана: %v\nAmount: %v\nProfit Margin: %v\nExpecting income: %v",
						 pot.BuyExchange, pot.SellExchange, pot.BuyPrice, pot.SellPrice, pot.BuyAmount, pot.ProfitMargin, pot.BuyAmount * (1 + pot.ProfitMargin))
					if _, err := bot.Send(tgbotapi.NewMessage(chatID, msgAS)); err != nil {
						logger.Log.Errorf("failed to send the message:%v", err)
						continue 
					}
				}


				time.Sleep(time.Second * 10)

			}
		}
	}()
	return nil
}

func StopAnalysis(chatID int64) error {
	tasksMU.Lock()
	defer tasksMU.Unlock()
	cancel, ok := RunningTasks[chatID]
	if !ok {
		logger.Log.Error("nothing to stop")
		return errors.New("no task running for this user")
	}
	cancel() 
	delete(RunningTasks, chatID)
	logger.Log.Infof("Analysis stopped for chatID %d", chatID)
	return nil
}