package db

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// Open creates a SQLite connection pool whose every connection enables
// foreign-key enforcement through the modernc.org/sqlite DSN pragma.
func Open(path string) (*sql.DB, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("database path is empty")
	}

	dsn, err := dataSourceName(path)
	if err != nil {
		return nil, fmt.Errorf("resolve database path: %w", err)
	}

	database, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	if err := database.PingContext(context.Background()); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("ping sqlite database: %w", err)
	}

	var foreignKeys int
	if err := database.QueryRowContext(context.Background(), "PRAGMA foreign_keys").Scan(&foreignKeys); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("read sqlite foreign_keys setting: %w", err)
	}
	if foreignKeys != 1 {
		_ = database.Close()
		return nil, fmt.Errorf("sqlite foreign key enforcement is disabled")
	}

	return database, nil
}

func dataSourceName(path string) (string, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	uriPath := filepath.ToSlash(absolutePath)
	if filepath.VolumeName(absolutePath) != "" && !strings.HasPrefix(uriPath, "/") {
		uriPath = "/" + uriPath
	}

	databaseURL := url.URL{Scheme: "file", Path: uriPath}
	query := databaseURL.Query()
	query.Add("_pragma", "foreign_keys(1)")
	// Concurrent create transactions wait for the current SQLite writer.
	query.Add("_pragma", "busy_timeout(10000)")
	databaseURL.RawQuery = query.Encode()
	return databaseURL.String(), nil
}
