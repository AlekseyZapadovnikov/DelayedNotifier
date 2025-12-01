package notify

import "context"

func (s *Notifier) CreateNotify(ctx context.Context, msg string, dateTime int64) error {
	return s.repo.AddNotify(msg, dateTime)
}

func (s *Notifier) GetNotifyStatByID(ctx context.Context, id int64) error {
	return s.repo.GetNotifyID(id)
}

func (s *Notifier) DeleteNotifyByID(ctx context.Context, id int64) error {
	return s.repo.DeleteNotifyByID(id)
}