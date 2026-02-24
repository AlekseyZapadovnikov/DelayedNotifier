package cache

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"time"

	"github.com/AlekseyZapadovnikov/DelayedNotifier/config"
	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/models"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

var (
	
)

type NotificationRepository interface {
	Save(ctx context.Context, rec *models.Record) (int64, error)
	GetByID(ctx context.Context, id int64) (*models.Record, error)
	DeleteByID(ctx context.Context, id int64) error
}

type ReddisCache struct {
	client    *redis.Client
	repo      NotificationRepository
	recordTTL time.Duration
}

func New(ctx context.Context, cfg config.RedisConfig, repo NotificationRepository) (*ReddisCache, error) {
	opts := &redis.Options{
		Addr:     cfg.Addr,
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.DB,

		Protocol: cfg.Protocol,

		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		PoolTimeout:  cfg.PoolTimeout,
		PoolSize:     cfg.PoolSize,

		ContextTimeoutEnabled: true,
	}

	if cfg.UseTLS {
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	rdb := redis.NewClient(opts)

	// Инструментирование (traces + metrics)
	if err := errors.Join(
		redisotel.InstrumentTracing(rdb),
		redisotel.InstrumentMetrics(rdb),
	); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("instrument redis: %w", err)
	}

	// Fail fast на старте
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := rdb.Ping(pingCtx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return &ReddisCache{
		client:    rdb,
		repo:      repo,
		recordTTL: cfg.RecordTTL,
	}, nil
}

func (c *ReddisCache) Close() error {
	return c.client.Close()
}
