package sender

import (
	"testing"
	"time"

	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/models"
	mailsender "github.com/AlekseyZapadovnikov/DelayedNotifier/internal/service/sender/mailsender"
	tgsender "github.com/AlekseyZapadovnikov/DelayedNotifier/internal/service/sender/tgsender"
)

func TestSenderInterface(t *testing.T) {
	// Проверяем что EmailSender реализует интерфейс Sender
	var _ Sender = (*mailsender.EmailSender)(nil)

	// Проверяем что TgSender реализует интерфейс Sender
	var _ Sender = (*tgsender.TgSender)(nil)
}

func TestEmailSenderGetType(t *testing.T) {
	es := &mailsender.EmailSender{}
	if es.GetType() != "mail" {
		t.Errorf("Expected EmailSender.GetType() to return 'mail', got %s", es.GetType())
	}
}

func TestTgSenderGetType(t *testing.T) {
	ts := &tgsender.TgSender{}
	if ts.GetType() != "tg" {
		t.Errorf("Expected TgSender.GetType() to return 'tg', got %s", ts.GetType())
	}
}

// Тест проверяет правильность работы с Record
func TestRecordValidation(t *testing.T) {
	// Создаем тестовый Record для email
	emailRecord := models.NewRecord(
		1,
		[]byte("Test email message"),
		time.Now(),
		"Test Subject",
		"mail",
		"test@example.com",
		[]string{"recipient@example.com"},
	)

	// Проверяем что record создан правильно
	if emailRecord.SendChan != "mail" {
		t.Errorf("Expected SendChan to be 'mail', got %s", emailRecord.SendChan)
	}

	// Создаем тестовый Record для telegram
	tgRecord := models.NewRecord(
		2,
		[]byte("Test telegram message"),
		time.Now(),
		"Test Subject",
		"tg",
		"testbot",
		[]string{"123456789"},
	)

	// Проверяем что record создан правильно
	if tgRecord.SendChan != "tg" {
		t.Errorf("Expected SendChan to be 'tg', got %s", tgRecord.SendChan)
	}
}
