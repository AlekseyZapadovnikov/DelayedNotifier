package config

import (
	"strings"
	"time"
)

type SMTPConfig struct {
	Host     string `env:"EMAIL_SMTP_HOST" validate:"omitempty,hostname|ip"`
	Port     string `env:"EMAIL_SMTP_PORT" validate:"omitempty,valid_port"`
	User     string `env:"EMAIL_SMTP_USER"`
	Password string `env:"EMAIL_SMTP_PASSWORD"`
}

type TelegramConfig struct {
	BotToken string `env:"TG_SENDER_BOT_TOKEN"`
}

func (s *SMTPConfig) IsConfigured() bool {
	return strings.TrimSpace(s.Host) != "" &&
		strings.TrimSpace(s.Port) != "" &&
		strings.TrimSpace(s.User) != ""
}

func (t *TelegramConfig) IsConfigured() bool {
	return strings.TrimSpace(t.BotToken) != ""
}

type NotificationConfig struct {
	RetrStr SendRetryLogic `validate:"dive"`
}

type SendRetryLogic struct {
	Attempts int           `env:"SEND_RETRY_ATTEMPTS" env-required:"true" validate:"min=1"`
	Delay    time.Duration `env:"SEND_RETRY_DELAY" env-required:"true" validate:"min=0"`
	Backoff  float64       `env:"SEND_RETRY_BACKOFF" env-required:"true" validate:"min=1"`
}
