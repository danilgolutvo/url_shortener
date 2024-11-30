package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgconn"
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

// SaveURL - save url and alias in database, checking if no alias with same name exists in DB if it does it show identifies it
func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const info = "storage.postgres.SaveURL"
	var id int64
	stmt := `INSERT INTO url(url, alias)
	VALUES ($1, $2) RETURNING id;`
	err := s.DB.QueryRow(context.Background(), stmt, urlToSave, alias).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, fmt.Errorf("%s: alias '%s' already exists: %w", info, alias, err)
		}
		return 0, fmt.Errorf("%s: %w", info, err)
	}
	return id, nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const info = "storage.postgres.GetURL"
	var url string
	stmt := `SELECT url FROM url WHERE alias = $1`
	err := s.DB.QueryRow(context.Background(), stmt, alias).Scan(&url)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("%s: No rows found for the given alias", info)
		}
		return "", fmt.Errorf("%s: %w", info, err)
	}
	return url, nil
}

func (s *Storage) DeleteURL(alias string) (bool, error) {
	const info = "storage.postgres.DeleteURL"
	ok, _ := s.CaseDifferent(alias)
	if ok {
		return ok, fmt.Errorf("no such alias, %s, %s", info, alias)
	}

	stmt := `DELETE FROM url WHERE alias = $1`
	_, err := s.DB.Exec(context.Background(), stmt, alias)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, fmt.Errorf("%s: No rows found for the given alias", info)
		}
		return false, fmt.Errorf("%s: %w", info, err)
	}
	return false, nil
}

func (s *Storage) CaseDifferent(alias string) (bool, error) {
	const info = "storage.postgres.Exists"
	stmt := `SELECT alias FROM url WHERE alias = $1`
	var aliasToCheck string
	err := s.DB.QueryRow(context.Background(), stmt, alias).Scan(&aliasToCheck)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, fmt.Errorf("%s: No rows found for the given alias", info)
		}
		return false, fmt.Errorf("%s: %w", info, err)
	}
	ok, _ := compareCaseSensitive(alias, aliasToCheck)
	return ok, nil
}

func compareCaseSensitive(str1, str2 string) (bool, error) {
	// Check if lengths are the same
	if len(str1) != len(str2) {
		return false, errors.New("strings have different lengths")
	}

	// Compare each character in both strings
	for i := 0; i < len(str1); i++ {
		if str1[i] != str2[i] {
			return false, fmt.Errorf("case mismatch at position %d: '%c' != '%c'", i, str1[i], str2[i])
		}
	}

	return true, nil
}
