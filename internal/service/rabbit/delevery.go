package rabbit

import (
	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/models"
)

type Delivery struct {
	rec      models.Record
	ackFunc  func() error
	nackFunc func(requeue bool) error
}

func (d *Delivery) Ack() error              { return d.ackFunc() }
func (d *Delivery) Nack(requeue bool) error { return d.nackFunc(requeue) }
func (d *Delivery) Record() *models.Record  { return &d.rec }