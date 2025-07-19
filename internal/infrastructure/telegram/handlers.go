package telegram

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Shyyw1e/arbitrage-sync/pkg/logger"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, store *UserStatesStore) error{
	chatID := msg.Chat.ID
	text := msg.Text

	if text == "/start" {
		store.Set(chatID, &UserState{
			Step: "waiting_for_input",
		})
		if _, err := bot.Send(tgbotapi.NewMessage(chatID, "Введите минимальную разницу и максимальную сумму через пробел. Например: 0.1 1000")); err != nil {
			logger.Log.Errorf("failed to send message: %v", err)
			return err
		}
		return nil
	}

	state, ok := store.Get(chatID)
	if !ok {
		if _, err := bot.Send(tgbotapi.NewMessage(chatID, "Сначала введите /start")); err != nil {
			logger.Log.Errorf("failed to send message: %v", err)
			return err
		}
		logger.Log.Error("invalid message")
		return errors.New("invalid message")
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
			return errors.New(fmt.Sprintf("%v\t%v", err1, err2))
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

func handleCallback(bot *tgbotapi.BotAPI, cb *tgbotapi.CallbackQuery, store *UserStatesStore) error {
	chatID := cb.Message.Chat.ID
	data := cb.Data
	userState, ok := store.Get(chatID)
	if !ok {
		logger.Log.Error("failed to get user state")
		if _, err := bot.Send(tgbotapi.NewMessage(chatID, "Сначала введите параметры.")); err != nil {
			logger.Log.Errorf("failed to send the message: %v", err)
			return err
		}
		return errors.New("no user state")
	}
	

	if data != "start_analysis" {
		logger.Log.Errorf("unexpected callback data: %s", data)
		return errors.New("unexpected callback data")
		
	}
	if userState.Step != "ready_to_run" {
		if _, err := bot.Send(tgbotapi.NewMessage(chatID, "Сначала введите параметры.")); err != nil {
			logger.Log.Errorf("failed to send the message: %v", err)
			return err
		}
		return errors.New("user not ready to run")
	}

	preMsg := fmt.Sprintf("Анализ запущен\nМинимальная разница: %v\nМаксимальная сумма: %v", userState.MinDiff, userState.MaxSum)
	if _, err := bot.Send(tgbotapi.NewMessage(chatID, preMsg)); err != nil{
		logger.Log.Errorf("failed to send the message: %v", err)
		return err
	}

	if _, err := bot.Request(tgbotapi.NewCallback(cb.ID, "Анализ запущен")); err != nil {
		logger.Log.Errorf("failed to send request: %v", err)
		return err
	}

	return nil
}