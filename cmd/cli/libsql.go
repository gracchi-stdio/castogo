package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gracchi-stdio/castogo/internal/repository"
	"github.com/gracchi-stdio/castogo/internal/repository/sqlite"
)

// sqliteDatabase is the dialect wrapper for every SQLite-flavored backend —
// remote libSQL (Bunny, over libsql-client-go) and local files (modernc
// driver). Both speak the same SQL dialect through database/sql.
type sqliteDatabase struct {
	db *sql.DB
}

// sqliteMigrations is the LibSQL/Bunny implementation of migrationHandler:
// TEXT timestamps via datetime('now'), ? placeholders, database/sql calls,
// DropAll via sqlite_master, Lock a no-op (writes serialize through the
// primary and schema_migrations's PK makes a double-apply fail safely).
type sqliteMigrations struct {
	db *sql.DB
}

func (d *sqliteDatabase) Ping(ctx context.Context) error { return d.db.PingContext(ctx) }

func (d *sqliteDatabase) Close() error {
	return d.db.Close()
}

func (d *sqliteDatabase) Migrations() migrationHandler { return &sqliteMigrations{db: d.db} }

func (d *sqliteDatabase) MigrationsDir() string { return filepath.Join("sql", "migrations", "sqlite") }

func (d *sqliteDatabase) UserRepo() repository.UserRepository {
	return sqlite.NewUserRepo(d.db)
}

func (d *sqliteDatabase) EpisodeRepo() repository.EpisodeRepository {
	return sqlite.NewEpisodeRepo(d.db)
}

func (m *sqliteMigrations) Lock(ctx context.Context) (func(), error) {
	return func() {}, nil
}

func (m *sqliteMigrations) EnsureMigrationsTable(ctx context.Context) error {
	_, err := m.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version TEXT PRIMARY KEY,
	applied_at TEXT NOT NULL DEFAULT (datetime('now'))
);
`)
	return err
}

func (m *sqliteMigrations) Applied(ctx context.Context) (map[string]struct{}, error) {
	rows, err := m.db.QueryContext(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("query schema_migrations: %w", err)
	}
	defer func() { _ = rows.Close() }()

	applied := make(map[string]struct{})
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("scan schema_migrations: %w", err)
		}
		applied[version] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate schema_migrations: %w", err)
	}
	return applied, nil
}

func (m *sqliteMigrations) Apply(ctx context.Context, dir, filename string) error {
	content, err := os.ReadFile(filepath.Join(dir, filename))
	if err != nil {
		return fmt.Errorf("read migration %s: %w", filename, err)
	}

	log.Printf("Applying migration %s", filename)
	if len(strings.TrimSpace(string(content))) == 0 {
		return fmt.Errorf("migration %s is empty", filename)
	}

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", filename, err)
	}

	if _, err := tx.ExecContext(ctx, string(content)); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("execute migration %s: %w", filename, err)
	}

	if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations(version) VALUES (?)", filename); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("record migration %s: %w", filename, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", filename, err)
	}
	return nil
}

func (m *sqliteMigrations) Record(ctx context.Context, filename string) error {
	_, err := m.db.ExecContext(ctx, "INSERT OR IGNORE INTO schema_migrations(version) VALUES (?)", filename)
	return err
}

func (m *sqliteMigrations) DropAll(ctx context.Context) error {
	rows, err := m.db.QueryContext(ctx, `SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%'`)
	if err != nil {
		return fmt.Errorf("list tables: %w", err)
	}
	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			_ = rows.Close()
			return fmt.Errorf("scan table name: %w", err)
		}
		tables = append(tables, name)
	}
	_ = rows.Close()
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate tables: %w", err)
	}

	for _, t := range tables {
		if _, err := m.db.ExecContext(ctx, fmt.Sprintf(`DROP TABLE IF EXISTS "%s"`, t)); err != nil {
			return fmt.Errorf("drop table %s: %w", t, err)
		}
	}
	return nil
}
