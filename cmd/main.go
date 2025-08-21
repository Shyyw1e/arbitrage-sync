package main

import (
	"os"

	"github.com/Shyyw1e/arbitrage-sync/internal/infrastructure/cache"
	"github.com/Shyyw1e/arbitrage-sync/internal/infrastructure/db"
	"github.com/Shyyw1e/arbitrage-sync/internal/infrastructure/parser"
	"github.com/Shyyw1e/arbitrage-sync/internal/infrastructure/redisqueue"
	"github.com/Shyyw1e/arbitrage-sync/internal/infrastructure/telegram"
	"github.com/Shyyw1e/arbitrage-sync/pkg/logger"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

var GlobalOrderCache *cache.OrderCache

func main() {
	logger.InitLog("debug")

	if err := godotenv.Load(); err != nil {
		logger.Log.Errorf("failed to load .env: %v", err)
	}

	if err := parser.StartChromeAllocator(); err != nil {
		logger.Log.Fatalf("chrome allocator start: %v", err)
	}
	defer parser.StopChromeAllocator()
	parser.SetChromeParallelLimit(1)

	if err := redisqueue.InitRedisClient(); err != nil {
		logger.Log.Fatalf("failed to init redis: %v", err)
	}

	store, err := db.NewSQLiteUserStateStore("data.db")
	if err != nil {
		logger.Log.Fatalf("failed to initialize SQLite store: %v", err)
	}

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		logger.Log.Fatalf("TELEGRAM_BOT_TOKEN is empty")
	}
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		logger.Log.Fatalf("Telegram bot init error: %v", err)
	}

	redisqueue.InitRedisQueue(store)
	go redisqueue.StartWorkerLoop(bot)

	telegram.StartBotWithBot(bot, store)
}
