package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

// NewRemoteDB opens a remote libSQL database over HTTP/WebSocket
// (libsql-client-go) — Bunny managed DB and other sqld servers.
func NewRemoteDB(ctx context.Context, databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("libsql", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open libsql: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return db, nil
}
