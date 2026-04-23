package session

import (
	"context"
	"encoding/base32"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PGStore implements gorilla/sessions.Store backed by PostgreSQL via pgx.
type PGStore struct {
	Codecs  []securecookie.Codec
	Options *sessions.Options
	Pool    *pgxpool.Pool
}

// NewPGStore creates a PGStore from an existing pgxpool.
func NewPGStore(pool *pgxpool.Pool, keyPairs ...[]byte) (*PGStore, error) {
	s := &PGStore{
		Codecs: securecookie.CodecsFromPairs(keyPairs...),
		Options: &sessions.Options{
			Path:   "/",
			MaxAge: 86400 * 30,
		},
		Pool: pool,
	}

	if err := s.createTable(); err != nil {
		return nil, err
	}

	return s, nil
}

// Get fetches a session for a given name after it has been added to the registry.
func (s *PGStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

// New returns a new session without adding it to the registry.
func (s *PGStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(s, name)
	opts := *s.Options
	session.Options = &opts
	session.IsNew = true

	if c, err := r.Cookie(name); err == nil {
		if err := securecookie.DecodeMulti(name, c.Value, &session.ID, s.Codecs...); err == nil {
			if err := s.load(session); err == nil {
				session.IsNew = false
			} else if !errors.Is(err, pgx.ErrNoRows) {
				return session, err
			}
		}
	}

	s.MaxAge(s.Options.MaxAge)
	return session, nil
}

// Save persists the session to PostgreSQL and sets the cookie.
func (s *PGStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	if session.Options.MaxAge < 0 {
		if err := s.destroy(session); err != nil {
			return err
		}
		http.SetCookie(w, sessions.NewCookie(session.Name(), "", session.Options))
		return nil
	}

	if session.ID == "" {
		session.ID = strings.TrimRight(
			base32.StdEncoding.EncodeToString(
				securecookie.GenerateRandomKey(32),
			), "=")
	}

	if err := s.save(session); err != nil {
		return err
	}

	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, s.Codecs...)
	if err != nil {
		return err
	}

	http.SetCookie(w, sessions.NewCookie(session.Name(), encoded, session.Options))
	return nil
}

// MaxAge sets the maximum age for the store and underlying securecookie codecs.
func (s *PGStore) MaxAge(age int) {
	s.Options.MaxAge = age
	for _, codec := range s.Codecs {
		if sc, ok := codec.(*securecookie.SecureCookie); ok {
			sc.MaxAge(age)
		}
	}
}

// --- internal helpers ---

func (s *PGStore) load(session *sessions.Session) error {
	var data string
	err := s.Pool.QueryRow(
		context.Background(), "SELECT data FROM http_sessions WHERE key = $1", session.ID,
	).Scan(&data)
	if err != nil {
		return fmt.Errorf("session load: %w", err)
	}

	return securecookie.DecodeMulti(session.Name(), data, &session.Values, s.Codecs...)
}

func (s *PGStore) save(session *sessions.Session) error {
	encoded, err := securecookie.EncodeMulti(session.Name(), session.Values, s.Codecs...)
	if err != nil {
		return err
	}

	expiresOn := time.Now().Add(time.Second * time.Duration(session.Options.MaxAge))
	ctx := context.Background()

	if session.IsNew {
		_, err = s.Pool.Exec(ctx,
			"INSERT INTO http_sessions (key, data, created_on, expires_on) VALUES ($1, $2, NOW(), $3)",
			session.ID, encoded, expiresOn,
		)
	} else {
		_, err = s.Pool.Exec(ctx,
			"UPDATE http_sessions SET data = $1, expires_on = $2 WHERE key = $3",
			encoded, expiresOn, session.ID,
		)
	}

	return err
}

func (s *PGStore) destroy(session *sessions.Session) error {
	_, err := s.Pool.Exec(context.Background(), "DELETE FROM http_sessions WHERE key = $1", session.ID)
	return err
}

func (s *PGStore) createTable() error {
	_, err := s.Pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS http_sessions (
			id       BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			key      TEXT NOT NULL UNIQUE,
			data     TEXT NOT NULL,
			created_on  TIMESTAMPTZ DEFAULT NOW(),
			expires_on  TIMESTAMPTZ NOT NULL
		);
		CREATE INDEX IF NOT EXISTS http_sessions_expiry_idx ON http_sessions (expires_on);
		CREATE INDEX IF NOT EXISTS http_sessions_key_idx ON http_sessions (key);
	`)
	return err
}
