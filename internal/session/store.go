package session

import "github.com/gorilla/sessions"

// SessionStore is the interface for persisting sessions.
// Implementations: PGStore (PostgreSQL), SQLiteStore (SQLite/libSQL over
// database/sql — local files and remote Bunny DB).
type SessionStore interface {
	sessions.Store // embeds Get, New, Save
}
