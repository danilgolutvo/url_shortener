package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // init pgx driver
	"url_shortener/internal/lib/logger/sl"
)

type Storage struct {
	DB *pgxpool.Pool
}

func NewStorage(ctx context.Context, dsn string) (*Storage, error) {
	pool, err := NewPgPool(ctx, dsn)
	if err != nil {
		return nil, err
	}
	storage := &Storage{DB: pool}
	// run init queries
	err = storage.initDatabase(ctx)
	if err != nil {
		pool.Close()
		return nil, err
	}
	return storage, nil
}

func NewPgPool(ctx context.Context, dsn string) (pool *pgxpool.Pool, err error) {
	const op = "storage.postgres.NewPgPool"
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	pool, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, sl.Err(err))
	}

	return pool, err
}

func (s *Storage) initDatabase(ctx context.Context) error {
	initQueries := []string{
		`CREATE TABLE IF NOT EXISTS url (
			id SERIAL PRIMARY KEY,
			alias TEXT NOT NULL UNIQUE,
			url TEXT NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);`,
	}
	for _, query := range initQueries {
		_, err := s.DB.Exec(ctx, query)
		if err != nil {
			return fmt.Errorf("%w: failed to exec query", err)
		}
	}
	return nil
}
