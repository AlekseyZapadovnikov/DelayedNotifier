package postgres

import (
	"context"
	"fmt"

	"github.com/AlekseyZapadovnikov/DelayedNotifier/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RecordRepo struct {
	pool *pgxpool.Pool
}

func NewRecordRepo(cfg *config.PostgresConfig) *RecordRepo {
	pool, err := NewPool(context.TODO(), cfg)
	if err != nil {
		panic(fmt.Sprintf("NewRecordRepo initialisation faild, err = %v", err))
	}
	return &RecordRepo{pool: pool}
}

func NewPool(ctx context.Context, config *config.PostgresConfig) (*pgxpool.Pool, error) {
	connString := config.ConnectionString()

	pgxConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}
	c := config.PoolConfig

	pgxConfig.MaxConns = int32(c.MaxConns)
	pgxConfig.MinConns = int32(c.MinConns)
	pgxConfig.MaxConnIdleTime = c.MaxConnIdleTime
	pgxConfig.MaxConnLifetime = c.MaxConnLifetime
	pgxConfig.HealthCheckPeriod = c.HealthCheckPeriod
	pgxConfig.ConnConfig.ConnectTimeout = c.ConnectionTimeout

	pool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil
}