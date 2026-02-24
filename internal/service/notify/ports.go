package notify

import (
	"context"

	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/models"
)

type MessageSender interface {
	SendMessage(record *models.Record) error
}

type NotificationCache interface {
	Add(ctx context.Context, rec *models.Record) error
	GetByID(ctx context.Context, id int64) (*models.Record, error)
	DeleteByID(ctx context.Context, id int64) error
}

type Delivery interface {
	Ack() error
	Nack(requeue bool) error
	Record() *models.Record
}

type DelayedQueue interface {
	Consume(ctx context.Context) (<-chan Delivery, error)
}

// DeleadQueue сохраняет совместимость с существующим именем с опечаткой.
type DeleadQueue = DelayedQueue
