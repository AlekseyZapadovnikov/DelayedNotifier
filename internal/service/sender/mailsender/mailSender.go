package sender

import (
	"strings"

	"github.com/AlekseyZapadovnikov/DelayedNotifier/config"
	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/models"
	"gopkg.in/gomail.v2"
)

type EmailSender struct {
	dialer gomail.Dialer
}

func NewEmailSander(conf config.Config) (*EmailSender, error) {
	
}

func (es *EmailSender) SendMessage(record models.Record) {
	es.dialer
}



func RecordToEmailMassage(record models.Record) *gomail.Message {

	m := gomail.NewMessage()
	m.SetHeader("From", record.From)
	m.SetHeader("To", record.To...)
	if strings.TrimSpace(record.Subject) != "" {
		m.SetHeader("Subject", record.Subject)
	}
	m.SetBody("text", string(record.Data)) // TODO проверить будет ли это работать с русским текстом

	return m
}