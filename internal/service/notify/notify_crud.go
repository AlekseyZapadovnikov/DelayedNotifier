package notify

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/models"
)

var ErrNotificationRecordNotFound = errors.New("notification record not found")

type delayedPublisher interface {
	PublishRecordDelayed(ctx context.Context, rec *models.Record, ttl time.Duration) error
}

func (n *Notifier) CreateNotify(ctx context.Context, rec *models.Record) error {
	if rec == nil {
		return fmt.Errorf("record is nil")
	}
	if n.cache == nil {
		return fmt.Errorf("notification cache is nil")
	}

	if rec.RecStat == "" {
		rec.RecStat = models.RecordStatusWaiting
	}

	if err := n.cache.Add(ctx, rec); err != nil {
		return err
	}

	pub, ok := n.dq.(delayedPublisher)
	if !ok || pub == nil {
		slog.Warn("notification created but delayed publisher is unavailable",
			"record_id", rec.Id,
			"send_chan", rec.SendChan,
			"send_time", rec.SendTime,
		)
		return nil
	}

	ttl := time.Until(rec.SendTime)
	if ttl < 0 {
		ttl = 0
	}

	if err := pub.PublishRecordDelayed(ctx, rec, ttl); err != nil {
		return fmt.Errorf("publish delayed notification: %w", err)
	}

	slog.Info("notification scheduled",
		"record_id", rec.Id,
		"send_chan", rec.SendChan,
		"send_time", rec.SendTime,
		"ttl_ms", ttl.Milliseconds(),
	)

	return nil
}

func (n *Notifier) GetNotifyStatByID(ctx context.Context, id int64) (models.RecordStatus, error) {
	rec, err := n.cache.GetByID(ctx, id)
	if err != nil {
		return "", err
	}
	if rec == nil {
		return "", ErrNotificationRecordNotFound
	}

	return rec.RecStat, nil
}

func (n *Notifier) DeleteNotifyByID(ctx context.Context, id int64) error {
	return n.cache.DeleteByID(ctx, id)
}
