package sqlite

import (
	"fmt"
	"time"
)

// sqliteTimeLayout matches datetime('now') exactly: UTC, second precision, no
// zone suffix. Timestamp comparisons in the queries are string comparisons, so
// everything this package writes must use this exact format.
const sqliteTimeLayout = "2006-01-02 15:04:05"

// timeToSQLite formats for storage. A wrong format here (e.g. RFC3339) does
// not error — it silently mis-sorts against datetime('now') in the queries.
func timeToSQLite(t time.Time) string {
	return t.UTC().Format(sqliteTimeLayout)
}

// timePtrToSQLite is timeToSQLite for nullable columns; nil stays SQL NULL.
func timePtrToSQLite(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := timeToSQLite(*t)
	return &s
}

// timeFromSQLite is the strict parse for NOT NULL columns: a value that fails
// to parse is a bug worth surfacing, not a zero time.
func timeFromSQLite(s string) (time.Time, error) {
	t, err := time.Parse(sqliteTimeLayout, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse timestamp %q: %w", s, err)
	}
	return t, nil
}

// timePtrFromSQLite is timeFromSQLite for nullable columns; NULL passes
// through as nil.
func timePtrFromSQLite(s *string) (*time.Time, error) {
	if s == nil {
		return nil, nil
	}
	t, err := timeFromSQLite(*s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// sqliteTime is the lenient parse: NULL or malformed becomes the zero time.
// Only for updated_at, where "never updated" and "unparseable" carry the
// same information.
func sqliteTime(s *string) time.Time {
	if s == nil {
		return time.Time{}
	}
	t, _ := time.Parse(sqliteTimeLayout, *s)
	return t
}

// orZero flattens a nullable storage field into the domain's plain value.
func orZero[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// toPtr boxes v. Go can't take the address of a conversion (&int64(x));
// a generic can.
func toPtr[T any](v T) *T {
	return &v
}
