package rabbit

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

type DelayedMessage struct {
	TTL time.Duration
	Rec models.Record
}

func (rc *RabbitClient) PublishRecordDelayed(_ context.Context, rec *models.Record, ttl time.Duration) error {
	if rec == nil {
		return fmt.Errorf("record is nil")
	}

	if ttl < 0 {
		ttl = 0
	}

	return rc.PublishDelayedMessage(DelayedMessage{
		TTL: ttl,
		Rec: *rec,
	})
}

func (rc *RabbitClient) PublishDelayedMessage(dm DelayedMessage) error {
	if !rc.IsReady() {
		return fmt.Errorf("connection unavailable")
	}

	rc.mu.Lock()
	ch := rc.ch
	rc.mu.Unlock()

	if ch == nil {
		return fmt.Errorf("channel is nil")
	}

	var expiration string
	if dm.TTL > 0 {
		expiration = strconv.FormatInt(dm.TTL.Milliseconds(), 10)
	}

	body, err := json.Marshal(dm.Rec)
	if err != nil {
		return fmt.Errorf("marshal record: %w", err)
	}

	return ch.PublishWithContext(
		context.Background(),
		"",
		rc.cfg.Queues.WaitQueue,
		false,
		false,
		amqp.Publishing{
			Headers: amqp.Table{
				"x-message-id": dm.Rec.Id,
				"x-from":       dm.Rec.From,
			},
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Expiration:   expiration,
			Body:         body,
		},
	)
}
