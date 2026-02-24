package notify

import (
	"context"
	"log/slog"

	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/models"
)

func (n *Notifier) TryToSendNotification(delivery Delivery) error {
	rec := delivery.Record()

	err := n.doRetry(func() error {
		return n.sender.SendMessage(rec)
	})
	if err != nil {
		nackErr := delivery.Nack(true)
		if nackErr != nil {
			slog.Error("failed to nack message",
				"message_id", recordID(rec),
				"from", recordFrom(rec),
				"error", nackErr,
			)
		}

		return err
	}

	ackErr := delivery.Ack()
	if ackErr != nil {
		slog.Error("failed to ack message",
			"message_id", recordID(rec),
			"from", recordFrom(rec),
			"error", ackErr,
		)

		return ackErr
	}

	return nil
}

func (n *Notifier) SendMessageWorker() error {
	return n.sendMessageWorker(context.Background())
}

func (n *Notifier) SendMessageWorkerContext(ctx context.Context) error {
	return n.sendMessageWorker(ctx)
}

func (n *Notifier) sendMessageWorker(ctx context.Context) error {
	if n == nil {
		return nil
	}
	if n.dq == nil {
		return nil
	}

	ch, err := n.dq.Consume(ctx)
	if err != nil {
		return err
	}

	slog.Info("notification worker started")

	for delivery := range ch {
		err := n.TryToSendNotification(delivery)
		if err != nil {
			rec := delivery.Record()
			slog.Error("failed to send notification",
				"message_id", recordID(rec),
				"from", recordFrom(rec),
				"error", err,
			)
		}
	}

	slog.Info("notification worker stopped")

	return nil
}

func recordID(rec *models.Record) int64 {
	if rec == nil {
		return 0
	}

	return rec.Id
}

func recordFrom(rec *models.Record) string {
	if rec == nil {
		return ""
	}

	return rec.From
}
