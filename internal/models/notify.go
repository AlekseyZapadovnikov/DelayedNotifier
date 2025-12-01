package models

import (
	"fmt"
	"time"

	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/valid"
	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func init() {
	Validate = valid.Validate
}

type Record struct {
	id       int64
	Data     []byte
	SendTime time.Time
	RecStat  string   `validate:"oneof=sended waiting redused"`
	SendChan string   `validate:"oneof=tg mail"`
	From     string   `validate:"from_field"`
	To       []string `validate:"to_field"`
	Subject  string   `validate:"len<200"`
}

func NewRecord(id int64, data []byte, sendTime time.Time, subject, sendChan, from string, to []string) *Record {
	return &Record{
		id:       id,
		Data:     data,
		SendTime: sendTime,
		RecStat:  RecordStatusWaiting,
		SendChan: sendChan,
		From:     from,
		To:       to,
		Subject:  subject,
	}
}

func (r *Record) SetStatus(newStatus string) error {
	err := Validate.Var(newStatus, "oneof=sended waiting redused")
	if err != nil {
		return fmt.Errorf("you try to set invalid status, status may be only one of (sended, waiting, redused)")
	}
	r.RecStat = newStatus
	return nil
}

const (
	RecordStatusSended  = "sended"
	RecordStatusWaiting = "waiting"
	RecordStatusRedused = "redused"
)


