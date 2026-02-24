package config

import (
	"fmt"
)

type MailHogConfig struct {
	SMTPHost string `env:"MAILHOG_SMTP_HOST" env-required:"true" validate:"required,hostname|ip"`
	SMTPPort string `env:"MAILHOG_SMTP_PORT" env-required:"true" validate:"required,port"`
	APIURL   string `env:"MAILHOG_API_URL" env-required:"true" validate:"required"`
}

func (c *MailHogConfig) MustLoad() {
	if c == nil {
		panic("mailhog config receiver is nil")
	}

	var cfg MailHogConfig
	if err := readEnvConfig(&cfg); err != nil {
		panic(fmt.Errorf("mailhog config loading failed: %w", err))
	}

	if err := Validate.Struct(cfg); err != nil {
		panic(fmt.Errorf("mailhog config validation failed: %w", err))
	}

	*c = cfg
}
