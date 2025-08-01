package main

import (
	"github.com/Shyyw1e/arbitrage-sync/internal/infrastructure/cache"
	"github.com/Shyyw1e/arbitrage-sync/internal/infrastructure/db"
	"github.com/Shyyw1e/arbitrage-sync/internal/infrastructure/telegram"
	"github.com/Shyyw1e/arbitrage-sync/pkg/logger"
	"github.com/joho/godotenv"
)

var GlobalOrderCache *cache.OrderCache

func main() {
	logger.InitLog("debug")
	if err := godotenv.Load(); err != nil {
		logger.Log.Errorf("failed to load .env: %v", err)
		return
	}
    store, err := db.NewSQLiteUserStateStore("data.db")
    if err != nil {
        logger.Log.Fatalf("failed to initialize SQLite store: %v", err)
    }

    telegram.StartBot(store)
}
