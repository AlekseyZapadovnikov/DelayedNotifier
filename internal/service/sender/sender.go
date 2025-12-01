package sender

import "github.com/AlekseyZapadovnikov/DelayedNotifier/internal/models"

// Sender определяет общий интерфейс для всех отправителей уведомлений
type Sender interface {
	// SendMessage отправляет сообщение через соответствующий канал
	SendMessage(record *models.Record) error

	// GetType возвращает тип отправителя (mail, tg)
	GetType() string
}
