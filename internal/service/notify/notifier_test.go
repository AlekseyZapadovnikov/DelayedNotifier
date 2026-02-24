package notify

import (
	"context"
	"errors"
	"testing"

	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wb-go/wbf/retry"
)

type fakeSender struct {
	calls []*models.Record
	send  func(record *models.Record) error
}

func (f *fakeSender) SendMessage(record *models.Record) error {
	f.calls = append(f.calls, record)

	if f.send != nil {
		return f.send(record)
	}

	return nil
}

type fakeDelivery struct {
	rec       *models.Record
	ackErr    error
	nackErr   error
	ackCalls  int
	nackCalls []bool
}

func (f *fakeDelivery) Ack() error {
	f.ackCalls++
	return f.ackErr
}

func (f *fakeDelivery) Nack(requeue bool) error {
	f.nackCalls = append(f.nackCalls, requeue)
	return f.nackErr
}

func (f *fakeDelivery) Record() *models.Record {
	return f.rec
}

type fakeDelayedQueue struct {
	consumeCalls int
	lastCtx      context.Context
	consumeFn    func(ctx context.Context) (<-chan Delivery, error)
}

func (f *fakeDelayedQueue) Consume(ctx context.Context) (<-chan Delivery, error) {
	f.consumeCalls++
	f.lastCtx = ctx

	if f.consumeFn != nil {
		return f.consumeFn(ctx)
	}

	ch := make(chan Delivery)
	close(ch)
	return ch, nil
}

func newTestNotifier(sender MessageSender, dq DelayedQueue) *Notifier {
	return &Notifier{
		sender: sender,
		dq:     dq,
		retryRunner: func(fn func() error, _ retry.Strategy) error {
			return fn()
		},
	}
}

func TestNotifier_TryToSendNotification_Success(t *testing.T) {
	t.Parallel()

	rec := &models.Record{Id: 1, From: "service@example.com"}
	sender := &fakeSender{}
	delivery := &fakeDelivery{rec: rec}

	n := newTestNotifier(sender, nil)

	err := n.TryToSendNotification(delivery)

	require.NoError(t, err)
	require.Len(t, sender.calls, 1)
	assert.Same(t, rec, sender.calls[0])
	assert.Equal(t, 1, delivery.ackCalls)
	assert.Empty(t, delivery.nackCalls)
}

func TestNotifier_TryToSendNotification_SendError(t *testing.T) {
	t.Parallel()

	sendErr := errors.New("send failed")
	rec := &models.Record{Id: 2, From: "service@example.com"}
	sender := &fakeSender{
		send: func(record *models.Record) error {
			return sendErr
		},
	}
	delivery := &fakeDelivery{rec: rec}

	n := newTestNotifier(sender, nil)

	err := n.TryToSendNotification(delivery)

	require.Error(t, err)
	assert.ErrorIs(t, err, sendErr)
	assert.Equal(t, 0, delivery.ackCalls)
	assert.Equal(t, []bool{true}, delivery.nackCalls)
}

func TestNotifier_TryToSendNotification_SendErrorAndNackError_ReturnsSendError(t *testing.T) {
	t.Parallel()

	sendErr := errors.New("send failed")
	nackErr := errors.New("nack failed")

	sender := &fakeSender{
		send: func(record *models.Record) error {
			return sendErr
		},
	}
	delivery := &fakeDelivery{
		rec:     &models.Record{Id: 3, From: "service@example.com"},
		nackErr: nackErr,
	}

	n := newTestNotifier(sender, nil)

	err := n.TryToSendNotification(delivery)

	require.Error(t, err)
	assert.Same(t, sendErr, err)
	assert.Equal(t, 0, delivery.ackCalls)
	assert.Equal(t, []bool{true}, delivery.nackCalls)
}

func TestNotifier_TryToSendNotification_AckError(t *testing.T) {
	t.Parallel()

	ackErr := errors.New("ack failed")

	sender := &fakeSender{}
	delivery := &fakeDelivery{
		rec:    &models.Record{Id: 4, From: "service@example.com"},
		ackErr: ackErr,
	}

	n := newTestNotifier(sender, nil)

	err := n.TryToSendNotification(delivery)

	require.Error(t, err)
	assert.ErrorIs(t, err, ackErr)
	assert.Equal(t, 1, delivery.ackCalls)
	assert.Empty(t, delivery.nackCalls)
}

func TestNotifier_sendMessageWorker_ConsumeError(t *testing.T) {
	t.Parallel()

	consumeErr := errors.New("consume failed")
	queue := &fakeDelayedQueue{
		consumeFn: func(ctx context.Context) (<-chan Delivery, error) {
			return nil, consumeErr
		},
	}

	n := newTestNotifier(&fakeSender{}, queue)

	err := n.sendMessageWorker(context.Background())

	require.Error(t, err)
	assert.ErrorIs(t, err, consumeErr)
	assert.Equal(t, 1, queue.consumeCalls)
}

func TestNotifier_sendMessageWorker_ContinuesAfterMessageError(t *testing.T) {
	t.Parallel()

	sendErr := errors.New("send failed for first message")
	sender := &fakeSender{
		send: func(record *models.Record) error {
			if record.Id == 10 {
				return sendErr
			}
			return nil
		},
	}

	first := &fakeDelivery{rec: &models.Record{Id: 10, From: "first@example.com"}}
	second := &fakeDelivery{rec: &models.Record{Id: 20, From: "second@example.com"}}

	queue := &fakeDelayedQueue{
		consumeFn: func(ctx context.Context) (<-chan Delivery, error) {
			ch := make(chan Delivery, 2)
			ch <- first
			ch <- second
			close(ch)
			return ch, nil
		},
	}

	n := newTestNotifier(sender, queue)

	err := n.sendMessageWorker(context.Background())

	require.NoError(t, err)
	require.Len(t, sender.calls, 2)
	assert.Equal(t, int64(10), sender.calls[0].Id)
	assert.Equal(t, int64(20), sender.calls[1].Id)

	assert.Equal(t, 0, first.ackCalls)
	assert.Equal(t, []bool{true}, first.nackCalls)

	assert.Equal(t, 1, second.ackCalls)
	assert.Empty(t, second.nackCalls)
}
