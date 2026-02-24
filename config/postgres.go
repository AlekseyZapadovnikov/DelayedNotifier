package config

import "time"

type PostgresConfig struct {
	Host       string     `env:"pgHost" env-required:"true" validate:"required,hostname|ip"`
	Port       string     `env:"pgPort" env-required:"true" validate:"required,port"`
	User       string     `env:"pgUser" env-required:"true" validate:"required"`
	Password   string     `env:"pgPassword" env-required:"true" validate:"required"`
	DBName     string     `env:"pgDBName" env-required:"true" validate:"required"`
	PoolConfig PoolConfig `validate:"required,dive"`
}

func (p *PostgresConfig) ConnectionString() string {
	return "postgres://" + p.User + ":" + p.Password + "@" + p.Host + ":" + p.Port + "/" + p.DBName + "?sslmode=disable"
}

type PoolConfig struct {
	MaxConns          int           `env:"pgMaxConns" env-required:"true" validate:"required,min=1"`
	MinConns          int           `env:"pgMinConns" env-required:"true" validate:"min=0"`
	MaxConnLifetime   time.Duration `env:"pgMaxConnLifetime" env-required:"true" validate:"required,min=0"`
	MaxConnIdleTime   time.Duration `env:"pgMaxConnIdleTime" env-required:"true" validate:"required,min=0"`
	HealthCheckPeriod time.Duration `env:"pgHealthCheckPeriod" env-required:"true" validate:"required,min=0"`
	ConnectionTimeout time.Duration `env:"pgConnectionTimeout" env-required:"true" validate:"required,min=0"`
}
