package rabbit

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (r *RabbitClient) setupClient(conn *amqp.Connection) error {
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("fail to create channel: %w", err)
	}

	// Safety: закрываем канал при ошибке настройки
	defer func() {
		if err != nil {
			_ = ch.Close()
		}
	}()

	log.Printf("Declaring DLX: %s", r.cfg.Queues.DLXName)
	err = ch.ExchangeDeclare(
		r.cfg.Queues.DLXName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("fail to declare dlx: %w", err)
	}

	log.Printf("Declaring Work Queue: %s", r.cfg.Queues.WorkQueue)
	workQ, err := ch.QueueDeclare(
		r.cfg.Queues.WorkQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("fail to declare work queue: %w", err)
	}

	err = ch.QueueBind(
		workQ.Name,
		r.cfg.Queues.DeadQueueKey,
		r.cfg.Queues.DLXName,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("couldn't bind work queue: %w", err)
	}

	log.Printf("Declaring Wait Queue: %s", r.cfg.Queues.WaitQueue)
	args := amqp.Table{
		"x-dead-letter-exchange":    r.cfg.Queues.DLXName,
		"x-dead-letter-routing-key": r.cfg.Queues.DeadQueueKey,
	}

	_, err = ch.QueueDeclare(
		r.cfg.Queues.WaitQueue,
		true,
		false,
		false,
		false,
		args,
	)
	if err != nil {
		return fmt.Errorf("fail to declare wait queue: %w", err)
	}

	r.changeChannel(ch)

	r.mu.Lock()
	r.isReady = true
	r.mu.Unlock()

	return nil
}