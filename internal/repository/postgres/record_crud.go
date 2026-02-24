package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/models"
	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/repository"
	"github.com/jackc/pgx/v5"
)

func (r *RecordRepo) Save(ctx context.Context, rec *models.Record) (int64, error) {
	const op = "postgres.RecordRepo.SaveRecord"

	query := `
		INSERT INTO records (
			data,
			send_time,
			rec_stat,
			send_chan,
			"from",
			"to",
			subject
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	var id int64
	err := r.pool.QueryRow(ctx, query,
		rec.Data,
		rec.SendTime,
		rec.RecStat,
		rec.SendChan,
		rec.From,
		rec.To,
		rec.Subject,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (r *RecordRepo) GetByID(ctx context.Context, id int64) (*models.Record, error) {
	const op = "postgres.RecordRepo.GetByID"

	query := `
		SELECT
			id,
			data,
			send_time,
			rec_stat,
			send_chan,
			"from",
			"to",
			subject
		FROM records
		WHERE id = $1
	`

	var rec models.Record
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&rec.Id,
		&rec.Data,
		&rec.SendTime,
		&rec.RecStat,
		&rec.SendChan,
		&rec.From,
		&rec.To,
		&rec.Subject,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, repository.ErrNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &rec, nil
}

func (r *RecordRepo) DeleteByID(ctx context.Context, id int64) error {
	const op = "postgres.RecordRepo.DeleteByID"

	result, err := r.pool.Exec(ctx, `DELETE FROM records WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, repository.ErrNotFound)
	}

	return nil
}
