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

func handleMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, store db.UserStatesStore) error {
	chatID := msg.Chat.ID
	text := msg.Text

	if text == "/start" || text == "⚙ Изменить параметры" {
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
		if err != nil || state.Step != "ready_to_run" {
			_, _ = bot.Send(tgbotapi.NewMessage(chatID, "Сначала введите параметры."))
			return nil
		}

		preMsg := fmt.Sprintf("Запускаю анализ!\nМинимальная разница: %v\nМаксимальная сумма: %v", state.MinDiff, state.MaxSum)
		bot.Send(tgbotapi.NewMessage(chatID, preMsg))

		err = scheduler.StartAnalysisForUser(bot, chatID, state)
		if err != nil {
			logger.Log.Errorf("failed to start analysis: %v", err)
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
		scheduler.StopAnalysis(chatID)
		bot.Send(tgbotapi.NewMessage(chatID, ""))
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
	logger.Log.Warnf("Unexpected callback: %s", cb.Data)
	_, _ = bot.Send(tgbotapi.NewMessage(chatID, "Неизвестная команда."))
	return nil
}
