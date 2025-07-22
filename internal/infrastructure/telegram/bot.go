package telegram

import (
	"os"

	
	"github.com/Shyyw1e/arbitrage-sync/pkg/logger"
	"github.com/Shyyw1e/arbitrage-sync/internal/infrastructure/db"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)



func StartBot(store db.UserStatesStore) {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		logger.Log.Errorf("failed to create start bot: %v", err)
	}
	bot.Debug = true
	logger.Log.Infof("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 10
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			if err := handleMessage(bot, update.Message, store); err != nil {
				logger.Log.Errorf("failed to handle message: %v", err)
			}
		} else if update.CallbackQuery != nil{
			if err := handleCallback(bot, update.CallbackQuery, store); err != nil {
				logger.Log.Errorf("failed to handle callback: %v", err)
			}
		}
	}

}


