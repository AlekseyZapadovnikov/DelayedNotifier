package config

import "time"

type RedisConfig struct {
	Addr     string `env:"REDIS_ADDR" env-required:"true" validate:"required"`
	Username string `env:"REDIS_USERNAME"`
	Password string `env:"REDIS_PASSWORD"`
	DB       int    `env:"REDIS_DB" env-required:"true" validate:"min=0"`

	RecordTTL time.Duration `env:"REDIS_RECORD_TTL" env-required:"true" validate:"required,min=0"`

	UseTLS bool `env:"REDIS_USE_TLS" env-required:"true"`

	DialTimeout  time.Duration `env:"REDIS_DIAL_TIMEOUT" env-required:"true" validate:"required,min=0"`
	ReadTimeout  time.Duration `env:"REDIS_READ_TIMEOUT" env-required:"true" validate:"required,min=0"`
	WriteTimeout time.Duration `env:"REDIS_WRITE_TIMEOUT" env-required:"true" validate:"required,min=0"`
	PoolTimeout  time.Duration `env:"REDIS_POOL_TIMEOUT" env-required:"true" validate:"required,min=0"`
	PoolSize     int           `env:"REDIS_POOL_SIZE" env-required:"true" validate:"min=1"`

	// RESP2/RESP3
	Protocol int `env:"REDIS_PROTOCOL" env-required:"true" validate:"oneof=2 3"` // 2 or 3
}
