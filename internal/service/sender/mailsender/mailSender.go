package mailsend

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/AlekseyZapadovnikov/DelayedNotifier/config"
	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/models"
	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/service/sender/serr"

	"gopkg.in/gomail.v2"
)

type EmailSender struct {
	dialer *gomail.Dialer
}

// NewEmailSender создает новый экземпляр Email отправителя
func NewEmailSender(cfg config.EmailSenderConfig) (*EmailSender, error) {
	smtpHost := strings.TrimSpace(cfg.SMTPHost)
	smtpPort := strings.TrimSpace(cfg.SMTPPort)
	smtpUser := strings.TrimSpace(cfg.SMTPUser)
	smtpPassword := cfg.SMTPPassword

	port, err := strconv.Atoi(smtpPort)
	if err != nil {
		return nil, fmt.Errorf("invalid smtpPort: %w", err)
	}

	dialer := gomail.NewDialer(smtpHost, port, smtpUser, smtpPassword)

	slog.Info("Email sender configured",
		"smtpHost", smtpHost,
		"smtpPort", smtpPort,
		"smtpUser", smtpUser)

	return &EmailSender{
		dialer: dialer,
	}, nil
}

// SendMessage отправляет email сообщение
func (es *EmailSender) SendMessage(record *models.Record) error {
	if record == nil {
		return fmt.Errorf("record cannot be nil")
	}

	if record.SendChan != "mail" {
		return fmt.Errorf("invalid send channel need: %s, got: %s, error = %w", "mail",
			record.SendChan,
			serr.ErrUnsupportedSendType,
		)
	}

	if es.dialer == nil {
		return fmt.Errorf("email dialer is not initialized")
	}

	// Создаем email сообщение
	recordForSend := *record
	if es.dialer != nil && strings.TrimSpace(es.dialer.Username) != "" {
		recordForSend.From = es.dialer.Username
	}
	m := es.recordToEmailMessage(&recordForSend)

	// Создаем соединение с SMTP сервером
	sender, err := es.dialer.Dial()
	if err != nil {
		slog.Error("Failed to dial SMTP server",
			"error", err,
			"smtpHost", es.dialer.Host,
			"smtpPort", es.dialer.Port)
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer sender.Close()

	// Отправляем сообщение
	if err := gomail.Send(sender, m); err != nil {
		slog.Error("Failed to send email",
			"error", err,
			"from", record.From,
			"to", record.To,
			"subject", record.Subject)
		return fmt.Errorf("failed to send email: %w", err)
	}

	slog.Info("Email sent successfully",
		"from", record.From,
		"to", record.To,
		"subject", record.Subject)

	return nil
}

// GetType возвращает тип отправителя
func (es *EmailSender) GetType() string {
	return "mail"
}

// recordToEmailMessage конвертирует Record в email сообщение
func (es *EmailSender) recordToEmailMessage(record *models.Record) *gomail.Message {
	m := gomail.NewMessage()

	// Устанавливаем отправителя
	if strings.TrimSpace(record.From) != "" {
		m.SetHeader("From", record.From)
	} else {
		// Если From не указан, используем пользователя SMTP
		if es.dialer != nil {
			m.SetHeader("From", es.dialer.Username)
		}
	}

	// Устанавливаем получателей
	if len(record.To) > 0 {
		m.SetHeader("To", record.To...)
	}

	// Устанавливаем тему
	if strings.TrimSpace(record.Subject) != "" {
		m.SetHeader("Subject", record.Subject)
	} else {
		m.SetHeader("Subject", "Notification")
	}

	// Устанавливаем тело письма
	body := string(record.Data)
	if strings.TrimSpace(body) == "" {
		body = "Empty message"
	}

	// Устанавливаем контент тип для поддержки русского текста
	m.SetBody("text/plain; charset=UTF-8", body)

	return m
}
