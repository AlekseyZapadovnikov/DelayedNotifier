package rabbit

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/models"
	notifyports "github.com/AlekseyZapadovnikov/DelayedNotifier/internal/service/notify"
)

// MessageWrapper инкапсулирует сообщение и логику подтверждения
type MessageWrapper struct {
	Message *DelayedMessage
	// Используем замыкания для скрытия реализации amqp.Delivery от потребителя
	ackFunc  func() error
	nackFunc func(requeue bool) error
}

func (mw *MessageWrapper) Ack() error              { return mw.ackFunc() }
func (mw *MessageWrapper) Nack(requeue bool) error { return mw.nackFunc(requeue) }
func (mw *MessageWrapper) Record() *models.Record {
	if mw == nil || mw.Message == nil {
		return nil
	}

	return &mw.Message.Rec
}

var _ notifyports.Delivery = (*MessageWrapper)(nil)
var _ notifyports.DelayedQueue = (*RabbitClient)(nil)

func (r *RabbitClient) Consume(ctx context.Context) (<-chan notifyports.Delivery, error) {
	src, err := r.ConsumeDelayedMessages(ctx)
	if err != nil {
		return nil, err
	}

	out := make(chan notifyports.Delivery)
	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case mw, ok := <-src:
				if !ok {
					return
				}
				if mw == nil {
					continue
				}

				select {
				case <-ctx.Done():
					return
				case out <- mw:
				}
			}
		}
	}()

	return out, nil
}

func (r *RabbitClient) ConsumeDelayedMessages(ctx context.Context) (<-chan *MessageWrapper, error) {
	r.mu.Lock()
	ch := r.ch
	r.mu.Unlock()

	if ch == nil {
		return nil, fmt.Errorf("channel is not initialized")
	}

	// PrefetchCount = 1 для равномерного распределения нагрузки
	if err := ch.Qos(1, 0, false); err != nil {
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	deliveries, err := ch.ConsumeWithContext(
		ctx,
		r.cfg.Queues.WorkQueue,
		"",    // consumer tag
		false, // auto-ack: false (обязательно для надежности)
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register consumer: %w", err)
	}

	out := make(chan *MessageWrapper)

	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-deliveries:
				if !ok {
					return
				}

				var rec models.Record
				// Десериализация тела сообщения в структуру Record
				if err := json.Unmarshal(d.Body, &rec); err != nil {
					log.Printf("failed to unmarshal record: %v", err)
					// Если сообщение "битое" (poison pill), удаляем его из очереди (requeue: false)
					_ = d.Nack(false, false)
					continue
				}

				msg := &DelayedMessage{
					Rec: rec,
					// TTL обычно берется из заголовков, если он нужен на стороне потребителя
					TTL: 0,
				}

				wrapper := &MessageWrapper{
					Message: msg,
					ackFunc: func() error {
						return d.Ack(false)
					},
					nackFunc: func(requeue bool) error {
						return d.Nack(false, requeue)
					},
				}

				select {
				case <-ctx.Done():
					return
				case out <- wrapper:
				}
			}
		}
	}()

	return out, nil
}
