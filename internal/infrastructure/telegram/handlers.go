package telegram

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Shyyw1e/arbitrage-sync/internal/core/domain"
	"github.com/Shyyw1e/arbitrage-sync/internal/infrastructure/db"
	"github.com/Shyyw1e/arbitrage-sync/internal/infrastructure/scheduler"
	"github.com/Shyyw1e/arbitrage-sync/pkg/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, store db.UserStatesStore) error{
	chatID := msg.Chat.ID
	text := msg.Text

	if text == "/start" {
		store.Set(chatID, &domain.UserState{
			Step: "waiting_for_input",
		})
		if _, err := bot.Send(tgbotapi.NewMessage(chatID, "Введите минимальную разницу и максимальную сумму через пробел. Например: 0.1 1000")); err != nil {
			logger.Log.Errorf("failed to send message: %v", err)
			return err
		}
		return nil
	}

	state, err := store.Get(chatID)
	if err != nil {
		if _, err := bot.Send(tgbotapi.NewMessage(chatID, "Сначала введите /start")); err != nil {
			logger.Log.Errorf("failed to send message: %v", err)
			return err
		}
		logger.Log.Errorf("failed to get user state: %v", err)
		return err
	}

	if state.Step == "waiting_for_input" {
		parts := strings.Split(text, " ")
		if len(parts) != 2 {
			if _, err := bot.Send(tgbotapi.NewMessage(chatID, "Неверный формат. Введите два числа через пробел.")); err != nil {
				logger.Log.Errorf("failed to send message: %v", err)
				return err
			}
			return errors.New("invalid input format")
		}
		minDiff, err1 := strconv.ParseFloat(parts[0], 64)
		maxSum, err2 := strconv.ParseFloat(parts[1], 64)
		if err1 != nil || err2 != nil {
			logger.Log.Errorf("failed to parse values: %v\t%v", err1, err2)
			if _, err := bot.Send(tgbotapi.NewMessage(chatID, "Ошибка в числах. Попробуйте ещё раз.")); err != nil {
				logger.Log.Errorf("failed to send message: %v", err)
				return err
			}
			return fmt.Errorf("%v\t%v", err1, err2)
		}

		state.MinDiff = minDiff
		state.MaxSum = maxSum
		state.Step = "ready_to_run"
		store.Set(chatID, state)

		msg := tgbotapi.NewMessage(chatID, "Параметры сохранены. Нажмите кнопку, чтобы начать анализ.")
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("▶️ Начать анализ", "start_analysis"),
			),
		)
		bot.Send(msg)

	}
	return nil
}



func handleCallback(bot *tgbotapi.BotAPI, cb *tgbotapi.CallbackQuery, store db.UserStatesStore) error {
	chatID := cb.Message.Chat.ID
	data := cb.Data

	userState, err := store.Get(chatID)
	if err != nil {
		logger.Log.Errorf("failed to get user state: %v", err)
		_, _ = bot.Send(tgbotapi.NewMessage(chatID, "Сначала введите параметры."))
		return err
	}

	switch data {
	case "start_analysis":
		if userState.Step != "ready_to_run" {
			_, _ = bot.Send(tgbotapi.NewMessage(chatID, "Сначала введите параметры."))
			return errors.New("user not ready to run")
		}

		preMsg := fmt.Sprintf("Запускаю анализ!\nМинимальная разница: %v\nМаксимальная сумма: %v", userState.MinDiff, userState.MaxSum)
		if _, err := bot.Send(tgbotapi.NewMessage(chatID, preMsg)); err != nil {
			logger.Log.Errorf("failed to send pre-message: %v", err)
			return err
		}

		err := scheduler.StartAnalysisForUser(bot, chatID, userState)
		if err != nil {
			logger.Log.Errorf("failed to start analysis: %v", err)
			_, _ = bot.Send(tgbotapi.NewMessage(chatID, "Не удалось запустить анализ."))
			return err
		}

		msg := tgbotapi.NewMessage(chatID, "Анализ запущен. Нажмите кнопку ниже, чтобы остановить:")
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("⏹ Остановить анализ", "stop_analysis"),
			),
		)
		bot.Send(msg)

		bot.Request(tgbotapi.NewCallback(cb.ID, "Анализ запущен"))

	case "stop_analysis":
		_, _ = bot.Send(tgbotapi.NewMessage(chatID, "⏹ Останавливаю анализ..."))

		err := scheduler.StopAnalysis(chatID)
		if err != nil {
			logger.Log.Errorf("failed to stop analysis: %v", err)
			_, _ = bot.Send(tgbotapi.NewMessage(chatID, "Не удалось остановить анализ. Возможно, он уже завершён."))
			return err
		}

		_, _ = bot.Send(tgbotapi.NewMessage(chatID, "✅ Анализ успешно остановлен."))
		bot.Request(tgbotapi.NewCallback(cb.ID, "Анализ остановлен"))

	default:
		logger.Log.Errorf("unexpected callback data: %s", data)
		return errors.New("unexpected callback data")
	}

	return nil
}
