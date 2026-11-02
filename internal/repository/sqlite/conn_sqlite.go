// Package sqlite is the SQLite/libSQL dialect layer: the repositories (which
// are transport-agnostic, running over database/sql) and the two connection
// constructors — NewLocalDB (modernc driver, local files) and NewRemoteDB
// (libsql-client-go, HTTP/WebSocket to sqld servers like Bunny).
package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// NewLocalDB opens a local SQLite file (modernc.org/sqlite, pure Go) for
// development and CLI testing.
func NewLocalDB(ctx context.Context, databaseURL string) (*sql.DB, error) {
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
