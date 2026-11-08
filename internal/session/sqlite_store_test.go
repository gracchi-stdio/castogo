package session

import (
	"database/sql"
	"net/http/httptest"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/gorilla/securecookie"
	_ "modernc.org/sqlite"
)

func newTestStore(t *testing.T) (*SQLiteStore, *sql.DB) {
	t.Helper()

	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "sessions.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	store, err := NewSQLiteStore(db,
		securecookie.GenerateRandomKey(32),
		securecookie.GenerateRandomKey(16),
	)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	return store, db
}

func TestSQLiteStoreRoundTrip(t *testing.T) {
	store, _ := newTestStore(t)

	// Save a fresh session (no cookie in the request → IsNew → INSERT).
	r := httptest.NewRequest("POST", "/login", nil)
	w := httptest.NewRecorder()

	s, err := store.New(r, "session")
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	s.Values["user_id"] = "42"
	if err := store.Save(r, w, s); err != nil {
		t.Fatalf("save: %v", err)
	}

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("want 1 set-cookie, got %d", len(cookies))
	}

	// Replay the cookie on a fresh request → registry Get → DB load.
	r2 := httptest.NewRequest("GET", "/admin", nil)
	r2.AddCookie(cookies[0])
	s2, err := store.Get(r2, "session")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if s2.IsNew {
		t.Fatalf("session should load as existing, got IsNew")
	}
	if got := s2.Values["user_id"]; got != "42" {
		t.Fatalf("want user_id=42, got %v", got)
	}

	// Destroy: MaxAge < 0 -> DELETE + cleared cookie.
	s2.Options.MaxAge = -1
	w2 := httptest.NewRecorder()
	if err := store.Save(r2, w2, s2); err != nil {
		t.Fatalf("destroy: %v", err)
	}

	var n int
	if err := store.DB.QueryRow("SELECT COUNT(*) FROM http_sessions").Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 0 {
		t.Fatalf("row should be deleted, %d remain", n)
	}
}

func TestSQLiteStoreTimestampFormat(t *testing.T) {
	store, _ := newTestStore(t)

	r := httptest.NewRequest("POST", "/login", nil)
	w := httptest.NewRecorder()
	s, _ := store.New(r, "session")
	s.Values["user_id"] = "42"
	if err := store.Save(r, w, s); err != nil {
		t.Fatalf("save: %v", err)
	}

	var expiresOn string
	if err := store.DB.QueryRow("SELECT expires_on FROM http_sessions").Scan(&expiresOn); err != nil {
		t.Fatalf("select: %v", err)
	}
	if !regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$`).MatchString(expiresOn) {
		t.Fatalf("expires_on %q does not match the datetime('now') layout", expiresOn)
	}
}
