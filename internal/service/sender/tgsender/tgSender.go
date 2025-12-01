package sender

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TgSender struct {
	bot *tgbotapi.BotAPI
}

// NewTgSender создает новый экземпляр Telegram отправителя
func NewTgSender(botToken string) (*TgSender, error) {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Telegram bot: %w", err)
	}

	slog.Info("Telegram bot authorized successfully", "username", bot.Self.UserName)

	return &TgSender{
		bot: bot,
	}, nil
}

// SendMessage отправляет сообщение в Telegram
func (ts *TgSender) SendMessage(record *models.Record) error {
	if record == nil {
		return fmt.Errorf("record cannot be nil")
	}

	if record.SendChan != "tg" {
		return fmt.Errorf("record is not intended for Telegram sending")
	}

	// Формируем текст сообщения
	messageText := ts.formatMessage(record)

	for _, recipient := range record.To {
		// Убираем символ @ в начале, если он есть
		chatIDStr := strings.TrimPrefix(recipient, "@")

		var msg tgbotapi.Chattable
		var err error

		// Пытаемся преобразовать в int64, если не получается - используем как username
		if strings.HasPrefix(chatIDStr, "-100") || (len(chatIDStr) > 7 && chatIDStr[0] != '-') {
			// Это ID чата (начинается с -100 для групп/каналов или длинный числовой ID)
			var chatID int64
			chatID, err = strconv.ParseInt(chatIDStr, 10, 64)
			if err != nil {
				slog.Error("Failed to parse chat ID as number", "chatIDStr", chatIDStr, "error", err)
				continue
			}
			msg = tgbotapi.NewMessage(chatID, messageText)
		} else {
			// Это username - для username нужно использовать специальный метод
			// В данной библиотеке для отправки по username сначала нужно получить chat ID
			// Пока пропускаем username и логируем предупреждение
			slog.Warn("Username sending not implemented yet, please use numeric chat ID", "username", chatIDStr)
			continue
		}

		// Форматируем сообщение с темой
		finalText := messageText
		if strings.TrimSpace(record.Subject) != "" {
			finalText = fmt.Sprintf("*%s*\n\n%s", record.Subject, messageText)
		}

		// Устанавливаем текст и форматирование
		switch m := msg.(type) {
		case tgbotapi.MessageConfig:
			m.Text = finalText
			m.ParseMode = "Markdown"
			msg = m
		}

		// Отправляем сообщение
		_, err = ts.bot.Send(msg)
		if err != nil {
			slog.Error("Failed to send Telegram message",
				"error", err,
				"recipient", recipient,
				"subject", record.Subject)
			return fmt.Errorf("failed to send message to %s: %w", recipient, err)
		}

		slog.Info("Telegram message sent successfully",
			"recipient", recipient,
			"subject", record.Subject)
	}

	return nil
}

// GetType возвращает тип отправителя
func (ts *TgSender) GetType() string {
	return "tg"
}

// formatMessage форматирует текст сообщения для отправки
func (ts *TgSender) formatMessage(record *models.Record) string {
	data := string(record.Data)

	// Если есть поле From, добавляем его в начало сообщения
	if strings.TrimSpace(record.From) != "" {
		return fmt.Sprintf("От: %s\n\n%s", record.From, data)
	}

	return data
}
