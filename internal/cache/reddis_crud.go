package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/models"
	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/repository"
	"github.com/redis/go-redis/v9"
)

func notyKey(id int64) string {
	return fmt.Sprintf("noty:%d", id)
}

func (rc *ReddisCache) Add(ctx context.Context, rec *models.Record) error {
	op := "cache.Add"
	if rec == nil {
		return fmt.Errorf("%s: record is nil", op)
	}

	id, err := rc.repo.Save(ctx, rec)
	if err != nil {
		return fmt.Errorf("%s: save to db: %w", op, err)
	}
	rec.Id = id

	key := notyKey(rec.Id)
	data, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("%s: marshal record: %w", op, err)
	}
	if err := rc.client.Set(ctx, key, data, rc.recordTTL).Err(); err != nil {
		return fmt.Errorf("%s: set to redis: %w", op, err)
	}

	return nil
}

func (rc *ReddisCache) GetByID(ctx context.Context, id int64) (*models.Record, error) {
	op := "cache.GetByID"
	key := notyKey(id)
	data, err := rc.client.Get(ctx, key).Bytes()
	if err == nil {
		var rec models.Record
		if err := json.Unmarshal(data, &rec); err != nil {
			_ = rc.client.Del(ctx, key).Err()
		} else {
			return &rec, nil
		}
	}

	if err != nil && !errors.Is(err, redis.Nil) {
		// Redis unavailable/corrupted entry: try DB as fallback.
		// Returning data is more important for API GET than strict cache usage.
	}

	rec, repoErr := rc.repo.GetByID(ctx, id)
	if repoErr != nil {
		if errors.Is(repoErr, repository.ErrNotFound) {
			return nil, nil
		}
		if err != nil && !errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("%s: redis err: %v; repo err: %w", op, err, repoErr)
		}
		return nil, fmt.Errorf("%s: get from db: %w", op, repoErr)
	}
	if rec == nil {
		return nil, nil
	}

	if data, marshalErr := json.Marshal(rec); marshalErr == nil {
		_ = rc.client.Set(ctx, key, data, rc.recordTTL).Err()
	}

	return rec, nil
}

func (rc *ReddisCache) DeleteByID(ctx context.Context, id int64) error {
	op := "cache.DeleteByID"
	key := notyKey(id)
	if err := rc.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("%s: delete from redis: %w", op, err)
	}

	if err := rc.repo.DeleteByID(ctx, id); err != nil {
		return fmt.Errorf("%s: delete from db: %w", op, err)
	}

	return nil
}
