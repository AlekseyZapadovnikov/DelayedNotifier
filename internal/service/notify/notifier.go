package notify

type NotifyRepository interface {
	AddNotify(msg string, dateTime int64) error
	GetNotifyID(id int64) error
	DeleteNotifyByID(id int64) error
}

type Notifier struct {
	repo NotifyRepository
}