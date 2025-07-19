package telegram

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Shyyw1e/arbitrage-sync/pkg/logger"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type UserState struct {
	MinDiff float64
	MaxSum   float64
	Step    string		//"waiting_foe_input", "ready_to_run", etc.
}


type UserStatesStore struct {
	mu sync.RWMutex
	store map[int64]*UserState
}

func NewUserStatesStore() *UserStatesStore {
	return &UserStatesStore{
		store: make(map[int64]*UserState),
	}
}

func (s *UserStatesStore) Get(chatID int64) (*UserState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, ok := s.store[chatID]
	return val, ok
}

func (s *UserStatesStore) Set(chatID int64, state *UserState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[chatID] = state
}

func (s *UserStatesStore) Delete(chatID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.store, chatID)
}

func StartBot(store *UserStatesStore) error{
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		logger.Log.Errorf("failed to create start bot: %v", err)
		return err
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
				return err
			}
		} else if update.CallbackQuery != nil{
			if err := handleCallback(bot, update.CallbackQuery, store); err != nil {
				logger.Log.Errorf("failed to handle callback: %v", err)
				return err
			}
		}
	}

	
	return nil
}