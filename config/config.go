package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/valid"
	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate = valid.Validate // валидатор в приложении

type Config struct {
	HTTPHost        string `validate:"required,hostname|ip"`
	HTTPPort        string `validate:"required,valid_port"` // кастомный валидатор
	StaticFilesPath string `validate:"required"`

	// SMTP
	SMTPHost     string `validate:"required,hostname|ip"`
	SMTPPort     string `validate:"required,valid_port"`
	SMTPUser     string `validate:"required"`
	SMTPPassword string `validate:"required"`

	// Telegram
	TgBotToken string `validate:"required,notempty"`
}

func LoadConfig() (*Config, error) {
	httpHost := os.Getenv("httpHost")
	if strings.TrimSpace(httpHost) == "" {
		return nil, fmt.Errorf("httpHost must be set, check .env file")
	}

	httpPort := os.Getenv("httpPort")
	if strings.TrimSpace(httpPort) == "" {
		return nil, fmt.Errorf("httpPort must be set, check .env file")
	}

	staticFilesPath := os.Getenv("staticFilesPath")
	if strings.TrimSpace(staticFilesPath) == "" {
		return nil, fmt.Errorf("staticFilesPath must be set, check .env file")
	}

	// SMTP конфигурация
	smtpHost := os.Getenv("smtpHost")
	smtpPort := os.Getenv("smtpPort")
	smtpUser := os.Getenv("smtpUser")
	smtpPassword := os.Getenv("smtpPassword")

	// Telegram конфигурация
	tgBotToken := os.Getenv("tgBotToken")

	return &Config{
		HTTPHost:        httpHost,
		HTTPPort:        httpPort,
		StaticFilesPath: staticFilesPath,
		SMTPHost:        smtpHost,
		SMTPPort:        smtpPort,
		SMTPUser:        smtpUser,
		SMTPPassword:    smtpPassword,
		TgBotToken:      tgBotToken,
	}, nil
}

func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%s", c.HTTPHost, c.HTTPPort)
}

func (c *Config) GetStaticFilesPath() string {
	return c.StaticFilesPath
}

// GetSMTPConfig возвращает SMTP конфигурацию
func (c *Config) GetSMTPConfig() (host, port, user, password string) {
	return c.SMTPHost, c.SMTPPort, c.SMTPUser, c.SMTPPassword
}

// GetTgBotToken возвращает токен Telegram бота
func (c *Config) GetTgBotToken() string {
	return c.TgBotToken
}

// IsEmailConfigured проверяет настроена ли email отправка
func (c *Config) IsEmailConfigured() bool {
	return strings.TrimSpace(c.SMTPHost) != "" &&
		strings.TrimSpace(c.SMTPPort) != "" &&
		strings.TrimSpace(c.SMTPUser) != "" &&
		strings.TrimSpace(c.SMTPPassword) != ""
}

// IsTelegramConfigured проверьте настроена ли Telegram отправка
func (c *Config) IsTelegramConfigured() bool {
	return strings.TrimSpace(c.TgBotToken) != ""
}
