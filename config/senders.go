package config

type EmailSenderConfig struct {
	SMTPHost     string `validate:"required,hostname|ip"`
	SMTPPort     string `validate:"required,valid_port"`
	SMTPUser     string `validate:"required"`
	SMTPPassword string
}

type TelegramSenderConfig struct {
	BotToken string
}
