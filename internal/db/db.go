package db

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// Open opens and verifies a Postgres database connection.
func Open(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, fmt.Errorf("ping database: %w; close database: %w", err, closeErr)
		}

		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}

// Migrate applies database migrations from migrationsPath.
func Migrate(databaseURL, migrationsPath string) error {
	migrationSourceURL, err := fileSourceURL(migrationsPath)
	if err != nil {
		return err
	}

	m, err := migrate.New(migrationSourceURL, databaseURL)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer func() {
		sourceErr, databaseErr := m.Close()
		if sourceErr != nil || databaseErr != nil {
			// migrate.Close only releases resources; migration success should not be
			// hidden by a close error.
			return
		}
	}()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}

func fileSourceURL(path string) (string, error) {
	if !filepath.IsAbs(path) {
		absolutePath, err := filepath.Abs(path)
		if err != nil {
			return "", fmt.Errorf("resolve migrations path: %w", err)
		}
		path = absolutePath
	}

	return (&url.URL{Scheme: "file", Path: path}).String(), nil
}
