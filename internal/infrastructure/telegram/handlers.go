package telegram

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Shyyw1e/arbitrage-sync/internal/core/domain"
	"github.com/Shyyw1e/arbitrage-sync/internal/infrastructure/db"
	"github.com/Shyyw1e/arbitrage-sync/internal/infrastructure/redisqueue"
	"github.com/Shyyw1e/arbitrage-sync/pkg/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, store db.UserStatesStore) error {
	chatID := msg.Chat.ID
	text := msg.Text

	if text == "/start" || text == "⚙ Изменить параметры" {
		_ = redisqueue.StopAnalysis(store, chatID)
		logger.Log.Infof("User %d reset parameters", chatID)

		store.Delete(chatID)
		store.Set(chatID, &domain.UserState{
			Step: "waiting_for_input",
		})

		msg := tgbotapi.NewMessage(chatID, "Введите минимальную разницу и максимальную сумму через пробел. Например: 0.1 1000")
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		if _, err := bot.Send(msg); err != nil {
			logger.Log.Errorf("failed to send message: %v", err)
			return err
		}
		return nil
	}

	if text == "▶️ Начать анализ" {
		state, err := store.Get(chatID)
		if err != nil || (state.Step != "ready_to_run" && state.Step != "not_active") {
			bot.Send(tgbotapi.NewMessage(chatID, "Сначала введите параметры."))
			logger.Log.Infof("User %d tried to start without valid state", chatID)
			return nil
		}

		logger.Log.Infof("User %d starting analysis (MinDiff: %.2f, MaxSum: %.2f)", chatID, state.MinDiff, state.MaxSum)

		preMsg := fmt.Sprintf("Запускаю анализ!\nМинимальная разница: %.2f\nМаксимальная сумма: %.2f", state.MinDiff, state.MaxSum)
		bot.Send(tgbotapi.NewMessage(chatID, preMsg))

		err = redisqueue.StartAnalysisForUser(bot, chatID, state)
		if err != nil {
			logger.Log.Errorf("failed to start analysis for user %d: %v", chatID, err)
			bot.Send(tgbotapi.NewMessage(chatID, "Не удалось запустить анализ."))
			return nil
		}

		msg := tgbotapi.NewMessage(chatID, "Анализ запущен.")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("⏹ Остановить анализ"),
				tgbotapi.NewKeyboardButton("⚙ Изменить параметры"),
			),
		)
		bot.Send(msg)
		return nil
	}

	if text == "⏹ Остановить анализ" {
		logger.Log.Infof("User %d requested analysis stop", chatID)

		err := redisqueue.StopAnalysis(store, chatID)
		if err != nil {
			logger.Log.Errorf("failed to stop analysis for user %d: %v", chatID, err)
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка при остановке анализа."))
			return err
		}

		msg := tgbotapi.NewMessage(chatID, "✅ Анализ успешно остановлен.")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("▶️ Начать анализ"),
				tgbotapi.NewKeyboardButton("⚙ Изменить параметры"),
			),
		)
		bot.Send(msg)
		return nil
	}

	state, err := store.Get(chatID)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Сначала введите /start"))
		logger.Log.Warnf("User %d sent message without state: %v", chatID, err)
		return err
	}

	if state.Step == "waiting_for_input" {
		parts := strings.Split(text, " ")
		if len(parts) != 2 {
			bot.Send(tgbotapi.NewMessage(chatID, "Неверный формат. Введите два числа через пробел."))
			return errors.New("invalid input format")
		}

		minDiff, err1 := strconv.ParseFloat(parts[0], 64)
		maxSum, err2 := strconv.ParseFloat(parts[1], 64)
		if err1 != nil || err2 != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "Ошибка в числах. Попробуйте ещё раз."))
			return fmt.Errorf("%v %v", err1, err2)
		}

		state.MinDiff = minDiff
		state.MaxSum = maxSum
		state.Step = "ready_to_run"
		store.Set(chatID, state)

		logger.Log.Infof("User %d set parameters: MinDiff = %.2f, MaxSum = %.2f", chatID, minDiff, maxSum)

		msg := tgbotapi.NewMessage(chatID, "Параметры сохранены. Нажмите кнопку, чтобы начать анализ.")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("▶️ Начать анализ"),
			),
		)
		bot.Send(msg)
		return nil
	}

	return nil
}


func handleCallback(bot *tgbotapi.BotAPI, cb *tgbotapi.CallbackQuery, store db.UserStatesStore) error {
	chatID := cb.Message.Chat.ID
	data := cb.Data

	logger.Log.Infof("Received callback from user %d: %s", chatID, data)

	switch data {
	default:
		logger.Log.Warnf("Unexpected callback data: %s", data)
		_, _ = bot.Send(tgbotapi.NewMessage(chatID, "Неизвестная команда."))
	}

	return nil
}