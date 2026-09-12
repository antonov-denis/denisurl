package store

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func (s *Store) Migrate(ctx context.Context) error {
	files, err := fs.Glob(migrationFS, "migrations/*.sql")
	if err != nil {
		return err
	}

	for _, name := range files {
		statements, err := migrationFS.ReadFile(name)
		if err != nil {
			return err
		}

		if _, err := s.pool.Exec(ctx, string(statements)); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}

	return nil
}
