package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // init pgx driver
	"golang.org/x/crypto/bcrypt"
	"time"
	"url_shortener/httpServer/handlers/login"
	"url_shortener/internal/lib/logger/sl"
	"url_shortener/internal/storage"
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
func (s *Storage) initDatabase(ctx context.Context) error {
	initQueries := []string{
		`CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,                      -- Use UUID for unique, standard IDs
    username TEXT NOT NULL UNIQUE,           -- Ensure usernames are unique
    password TEXT NOT NULL                   -- Password cannot be NULL
);`,
		`CREATE TABLE IF NOT EXISTS url (
	id UUID PRIMARY KEY,
	alias TEXT NOT NULL UNIQUE,
	url TEXT NOT NULL,
    creator UUID NOT NULL,                  -- Reference to the users table
    createdAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (creator) REFERENCES users(id) ON DELETE CASCADE);`,
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
func (s *Storage) CreateUser(user login.User) error {
	const info = "internal.storage.postgres.users.CreateUsers"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("%s: failed to hash pass: %w", info, err)
	}
	stmt := `INSERT INTO users(id, username, password)
	VALUES ($1, $2, $3);`

	_, err = s.DB.Exec(context.Background(), stmt, user.ID, user.Username, hashedPassword)
	if err != nil {
		return fmt.Errorf("%s: failed to insert new user: %w", info, err)
	}
	return nil
}

// SaveURL - save url and alias in database, checking if no alias with same name exists in DB if it does it show identifies it
func (s *Storage) SaveURL(urlToSave, alias, creator string) (string, error) {
	const info = "storage.postgres.SaveURL"
	id := uuid.New().String()
	var createdAt time.Time
	stmt := `INSERT INTO url(id, url, alias, creator)
	VALUES ($1, $2, $3, $4) RETURNING createdAt;`
	err := s.DB.QueryRow(context.Background(), stmt, id, urlToSave, alias, creator).Scan(&createdAt)
	if err != nil {
		return "", fmt.Errorf("%s: failed to insert entry: %w", info, err)
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

// DeleteURL deletes a URL identified by the alias and creator from the database.
func (s *Storage) DeleteURL(alias, creator string) (bool, error) {
	const info = "storage.postgres.DeleteURL"

	// Verify the alias exists and is case-sensitive
	ok, err := s.CaseDifferent(alias)
	if err != nil {
		return false, fmt.Errorf("%s: %w", info, err)
	}
	if !ok {
		return false, fmt.Errorf("%s: %s, %w", info, alias, storage.ErrCaseMismatch)
	}

	// Prepare and execute the delete statement
	stmt := `DELETE FROM url WHERE alias = $1 AND creator = $2`
	result, err := s.DB.Exec(context.Background(), stmt, alias, creator)
	if err != nil {
		return false, fmt.Errorf("%s: failed to execute delete statement: %w", info, err)
	}

	// Check if rows were actually affected
	if result.RowsAffected() == 0 {
		return false, fmt.Errorf("%s: no rows found for alias %s and creator %s, %w", info, alias, creator, storage.ErrAliasNotFound)
	}

	return true, nil
}

// CaseDifferent checks if the alias exists in a case-sensitive manner.
func (s *Storage) CaseDifferent(alias string) (bool, error) {
	const info = "storage.postgres.CaseDifferent"
	stmt := `SELECT alias FROM url WHERE alias = $1`

	var foundAlias string
	err := s.DB.QueryRow(context.Background(), stmt, alias).Scan(&foundAlias)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil // Alias does not exist, not an error
		}
		return false, fmt.Errorf("%s: query failed: %w", info, err)
	}

	// Perform case-sensitive comparison
	if alias == foundAlias {
		return true, nil
	}

	return false, nil
}
