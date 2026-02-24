package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/valid"
	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

var Validate *validator.Validate = valid.Validate

type Config struct {
	HTTP                 HTTPConfig           `validate:"required,dive"`
	Rabbit               RabbitConfig         `validate:"required,dive"`
	SMTP                 SMTPConfig           `validate:"dive"`
	Telegram             TelegramConfig       `validate:"dive"`
	Notifier             NotificationConfig   `validate:"required,dive"`
	PostgresConfig       PostgresConfig       `validate:"required,dive"`
	RedisConfig          RedisConfig          `validate:"required,dive"`
	EmailSenderConfig    EmailSenderConfig    `validate:"required,dive"`
	TelegramSenderConfig TelegramSenderConfig `validate:"required,dive"`
}

func LoadConfig() (*Config, error) {
	var cfg Config

	if err := readEnvConfig(&cfg); err != nil {
		return nil, fmt.Errorf("config loading failed: %w", err)
	}

	cfg.EmailSenderConfig = EmailSenderConfig{
		SMTPHost:     cfg.SMTP.Host,
		SMTPPort:     cfg.SMTP.Port,
		SMTPUser:     cfg.SMTP.User,
		SMTPPassword: cfg.SMTP.Password,
	}
	cfg.TelegramSenderConfig = TelegramSenderConfig{
		BotToken: cfg.Telegram.BotToken,
	}

	if err := Validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func readEnvConfig(dst any) error {
	const envFile = ".env"

	if _, err := os.Stat(envFile); err == nil {
		return cleanenv.ReadConfig(envFile, dst)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat %s: %w", envFile, err)
	}

	return cleanenv.ReadEnv(dst)
}
