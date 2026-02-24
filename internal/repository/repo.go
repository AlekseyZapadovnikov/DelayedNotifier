package repository

import (
	"context"
	"errors"

	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/models"
)

var (
	ErrNotFound      = errors.New("record not found")
	ErrAlreadyExists = errors.New("record already exists")
)

type NotificationRepository interface {
	Save(ctx context.Context, rec *models.Record) (int64, error)
	GetByID(ctx context.Context, id int64) (*models.Record, error)
	DeleteByID(ctx context.Context, id int64) error
}
