package store

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

var ErrNotFound = errors.New("link not found")

func (s *Store) GetTarget(ctx context.Context, code string) (string, error) {
	var target string
	err := s.pool.QueryRow(ctx, "SELECT target_url FROM links WHERE code = $1", code).Scan(&target)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}

		slog.Error(err.Error())
		return "", err
	}
	return target, nil
}

func (s *Store) CreateTarget(ctx context.Context, code string, target string) error {
	_, err := s.pool.Exec(ctx, "INSERT INTO links (code, target_url) VALUES ($1, $2)", code, target)
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	return nil
}

func (s *Store) Close() {
	s.pool.Close()
}

func New(databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	return &Store{pool: pool}, nil
}
