package config

import (
	"fmt"
	"time"
)

type RabbitConfig struct {
	Connection RabbitConnection        `validate:"required,dive"`
	Client     RabbitClientParams      `validate:"required,dive"`
	Retry      RabbitRetryPublishLogic `validate:"required,dive"`
	Queues     RabbitTopology          `validate:"required,dive"`
}

type RabbitTopology struct {
	WaitQueue    string `env:"AMQP_WAIT_QUEUE" env-required:"true" validate:"required"`
	WorkQueue    string `env:"AMQP_WORK_QUEUE" env-required:"true" validate:"required"`
	DLXName      string `env:"AMQP_DLX_NAME" env-required:"true" validate:"required"`
	DeadQueueKey string `env:"AMQP_DEAD_ROUTING_KEY" env-required:"true" validate:"required"`
}

type RabbitConnection struct {
	Host     string `env:"AMQP_HOST" env-required:"true" validate:"required,hostname|ip"`
	Port     string `env:"AMQP_PORT" env-required:"true" validate:"required,port"`
	User     string `env:"AMQP_USER" env-required:"true" validate:"required"`
	Password string `env:"AMQP_PASS" env-required:"true" validate:"required"`
	VHost    string `env:"AMQP_VHOST" env-required:"true" validate:"required"`
}

type RabbitClientParams struct {
	ServiceName    string        `env:"AMQP_CONNECTION_NAME" env-required:"true" validate:"required"`
	ConnectTimeout time.Duration `env:"AMQP_CONNECT_TIMEOUT" env-required:"true" validate:"required"`
	ReinitTimeout  time.Duration `env:"AMQP_REINIT_TIMEOUT" env-required:"true" validate:"required"`
	Heartbeat      time.Duration `env:"AMQP_HEARTBEAT" env-required:"true" validate:"required"`
}

type RabbitRetryPublishLogic struct {
	PublishAttempts int           `env:"PUBLISH_RETRY_ATTEMPTS" env-required:"true" validate:"required,min=1"`
	PublishDelay    time.Duration `env:"PUBLISH_RETRY_DELAY" env-required:"true" validate:"required,min=0"`
	PublishBackoff  float64       `env:"PUBLISH_RETRY_BACKOFF" env-required:"true" validate:"required,min=1.0"`
}

func (r *RabbitConnection) ConnectionString() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s%s",
		r.User, r.Password, r.Host, r.Port, r.VHost,
	)
}
