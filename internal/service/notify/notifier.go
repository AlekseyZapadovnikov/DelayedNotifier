package notify

import (
	"github.com/AlekseyZapadovnikov/DelayedNotifier/config"
	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/service/sender"
	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/web"

	"github.com/wb-go/wbf/retry"
)

type Notifier struct {
	curStrategy retry.Strategy
	sender      MessageSender
	cache       NotificationCache
	dq          DelayedQueue
	retryRunner func(fn func() error, strategy retry.Strategy) error
}

func NewNotifier(senderWrap sender.SenderWrap, cache NotificationCache, dq DelayedQueue, conf config.NotificationConfig) *Notifier {
	return NewNotifierWithSender(&senderWrap, cache, dq, conf)
}

func NewNotifierWithSender(sender MessageSender, cache NotificationCache, dq DelayedQueue, conf config.NotificationConfig) *Notifier {
	cs := retry.Strategy(conf.RetrStr)

	return &Notifier{
		curStrategy: cs,
		sender:      sender,
		cache:       cache,
		dq:          dq,
		retryRunner: retry.Do,
	}
}

var _ web.NotifydService = (*Notifier)(nil)

func (n *Notifier) doRetry(fn func() error) error {
	if n.retryRunner != nil {
		return n.retryRunner(fn, n.curStrategy)
	}

	return retry.Do(fn, n.curStrategy)
}
