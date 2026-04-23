package session

import "github.com/gorilla/sessions"

// SessionStore is the interface for persisting sessions.
// Implementations: PGStore (PostgreSQL), future LibSQLStore, RedisStore, etc.
type SessionStore interface {
	sessions.Store // embeds Get, New, Save
}
