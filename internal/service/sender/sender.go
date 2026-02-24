package sender

import (
	"fmt"

	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/models"
)

var (
	ErrUnsupportedSendType = fmt.Errorf("unsupported send type")
)

// Sender определяет общий интерфейс для всех отправителей уведомлений
type Sender interface {
	// SendMessage отправляет сообщение через соответствующий канал
	SendMessage(record *models.Record) error

	// GetType возвращает тип отправителя (mail, tg)
	GetType() string
}

type SenderWrap struct {
	tgSender   Sender
	mailSender Sender
}

func NewSenderWrap(tg Sender, mail Sender) SenderWrap {
	return SenderWrap{
		tgSender:   tg,
		mailSender: mail,
	}
}
func (sw *SenderWrap) SendMessage(record *models.Record) error {
	if sw == nil {
		return fmt.Errorf("sender wrapper is nil")
	}

	sendType := record.SendChan
	switch sendType {
	case models.SendChanTG:
		if sw.tgSender == nil {
			return fmt.Errorf("telegram sender is not configured: %w", ErrUnsupportedSendType)
		}
		return sw.tgSender.SendMessage(record)
	case models.SendChanMail:
		if sw.mailSender == nil {
			return fmt.Errorf("mail sender is not configured: %w", ErrUnsupportedSendType)
		}
		return sw.mailSender.SendMessage(record)
	default:
		return fmt.Errorf("unsupported send type: %s", sendType)
	}
}
