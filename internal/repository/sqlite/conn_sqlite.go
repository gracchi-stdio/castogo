// Package sqlite provides local-file SQLite connections (modernc.org/sqlite,
// pure Go) for development and CLI testing. Remote libSQL (Bunny) goes through
// the libsql package instead — same SQL dialect, different transport.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func NewDB(ctx context.Context, databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return db, nil
}
