package session

import (
	"context"
	"database/sql"
	"encoding/base32"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

// sqliteTimeLayout matches datetime('now') exactly: UTC, second precision,
// no zone suffix — the same convention as internal/repository/sqlite.
// expires_on is TEXT, so what this store writes must line up with
// datetime('now') for any future expiry-based cleanup to compare correctly.
const sqliteTimeLayout = "2006-01-02 15:04:05"

var _ SessionStore = (*SQLiteStore)(nil)

// SQLiteStore implements gorilla/sessions.Store over database/sql. One store
// serves both SQLite flavors in this codebase — local files (modernc driver)
// and remote libSQL (libsql-client-go, Bunny DB) — the driver is chosen by
// whoever opens the *sql.DB.
type SQLiteStore struct {
	Codecs  []securecookie.Codec
	Options *sessions.Options
	DB      *sql.DB
}

// Get implements [SessionStore].
func (s *SQLiteStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

// New implements [SessionStore].
func (s *SQLiteStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(s, name)
	opts := *s.Options
	session.Options = &opts
	session.IsNew = true

	if c, err := r.Cookie(name); err == nil {
		if err := securecookie.DecodeMulti(name, c.Value, &session.ID, s.Codecs...); err == nil {
			if err := s.load(r.Context(), session); err == nil {
				session.IsNew = false
			} else if !errors.Is(err, sql.ErrNoRows) {
				return session, err
			}
		}
	}

	s.MaxAge(s.Options.MaxAge)
	return session, nil
}

// MaxAge sets the maximum age for the store and underlying securecookie codecs.
func (s *SQLiteStore) MaxAge(age int) {
	s.Options.MaxAge = age
	for _, codec := range s.Codecs {
		if sc, ok := codec.(*securecookie.SecureCookie); ok {
			sc.MaxAge(age)
		}
	}
}

func (s *SQLiteStore) load(ctx context.Context, session *sessions.Session) error {
	var data string
	err := s.DB.QueryRowContext(
		ctx, "SELECT data FROM http_sessions WHERE key = ?", session.ID,
	).Scan(&data)
	if err != nil {
		return fmt.Errorf("session load: %w", err)
	}
	return securecookie.DecodeMulti(session.Name(), data, &session.Values, s.Codecs...)

}

// Save implements [SessionStore].
func (s *SQLiteStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	ctx := r.Context()

	if session.Options.MaxAge < 0 {
		if err := s.destroy(ctx, session); err != nil {
			return err
		}
		http.SetCookie(w, sessions.NewCookie(session.Name(), "", session.Options))
		return nil
	}

	if session.ID == "" {
		session.ID = strings.TrimRight(base32.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32)), "=")
	}

	if err := s.save(ctx, session); err != nil {
		return err
	}

	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, s.Codecs...)
	if err != nil {
		return err
	}
	http.SetCookie(w, sessions.NewCookie(session.Name(), encoded, session.Options))
	return nil

}

func (s *SQLiteStore) save(ctx context.Context, session *sessions.Session) error {
	encoded, err := securecookie.EncodeMulti(session.Name(), session.Values, s.Codecs...)
	if err != nil {
		return err
	}

	expiresOn := time.Now().Add(time.Second * time.Duration(session.Options.MaxAge)).
		UTC().Format(sqliteTimeLayout)

	if session.IsNew {
		_, err = s.DB.ExecContext(ctx,
			"INSERT INTO http_sessions (key, data, created_on, expires_on) VALUES (?, ?, datetime('now'), ?)",
			session.ID, encoded, expiresOn,
		)
	} else {
		_, err = s.DB.ExecContext(ctx,
			"UPDATE http_sessions SET data = ?, expires_on = ? WHERE key = ?",
			encoded, expiresOn, session.ID,
		)
	}

	return err
}

func (s *SQLiteStore) destroy(ctx context.Context, session *sessions.Session) error {
	_, err := s.DB.ExecContext(ctx, "DELETE FROM http_sessions WHERE key = ?", session.ID)
	return err
}

// NewSQLiteStore creates a SQLiteStore from an existing *sql.DB.
func NewSQLiteStore(db *sql.DB, keyPairs ...[]byte) (*SQLiteStore, error) {
	s := &SQLiteStore{
		Codecs: securecookie.CodecsFromPairs(keyPairs...),
		Options: &sessions.Options{
			Path:   "/",
			MaxAge: 86400 * 30,
		},
		DB: db,
	}

	if err := s.createTable(); err != nil {
		return nil, err
	}

	return s, nil

}

func (s *SQLiteStore) createTable() error {
	_, err := s.DB.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS http_sessions (
			id         INTEGER PRIMARY KEY,
			key        TEXT NOT NULL UNIQUE,
			data       TEXT NOT NULL,
			created_on TEXT DEFAULT (datetime('now')),
			expires_on TEXT NOT NULL
		);
		CREATE INDEX IF NOT EXISTS http_sessions_expiry_idx ON http_sessions (expires_on);
		CREATE INDEX IF NOT EXISTS http_sessions_key_idx ON http_sessions (key);
	`)
	return err
}
