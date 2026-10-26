package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type pgDatabase struct {
	pool *pgxpool.Pool
}

func (d *pgDatabase) Ping(ctx context.Context) error { return d.pool.Ping(ctx) }

func (d *pgDatabase) Close() error {
	d.pool.Close()
	return nil
}

func (d *pgDatabase) Migrations() migrationHandler { return &pgMigrations{pool: d.pool} }

func (d *pgDatabase) MigrationsDir() string { return filepath.Join("sql", "migrations") }

// pgMigrations is the PostgreSQL implementation of migrationHandler.
type pgMigrations struct {
	pool *pgxpool.Pool
}

// migrationsLockKey identifies castogo's migration advisory lock. Advisory
// locks live in the server's global lock space shared by every connection,
// so a fixed arbitrary constant means every migrator (deploy boot, manual CLI
// run) contends on the same key.
const migrationsLockKey = 0x63617374 // "cast" in hex — arbitrary but stable

const migrationsLockTimeout = 30 * time.Second

// Lock takes a session-level advisory lock on a dedicated pooled connection
// so that two concurrent migrators serialize instead of racing on
// schema_migrations (e.g. a redeploy boot-migrating while someone runs the
// CLI by hand against the same database).
func (m *pgMigrations) Lock(ctx context.Context) (func(), error) {
	conn, err := m.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire connection: %w", err)
	}

	deadline := time.Now().Add(migrationsLockTimeout)
	for {
		var locked bool
		if err := conn.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", migrationsLockKey).Scan(&locked); err != nil {
			conn.Release()
			return nil, fmt.Errorf("try advisory lock: %w", err)
		}
		if locked {
			break
		}
		if time.Now().After(deadline) {
			conn.Release()
			return nil, fmt.Errorf("another migration is in progress (lock busy for %s) — retry after it finishes", migrationsLockTimeout)
		}
		select {
		case <-ctx.Done():
			conn.Release()
			return nil, ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}

	release := func() {
		// Advisory locks are session-scoped: unlock on the same connection
		// before returning it to the pool.
		_, _ = conn.Exec(context.Background(), "SELECT pg_advisory_unlock($1)", migrationsLockKey)
		conn.Release()
	}
	return release, nil
}

func (m *pgMigrations) EnsureMigrationsTable(ctx context.Context) error {
	_, err := m.pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version TEXT PRIMARY KEY,
	applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`)
	return err
}

func (m *pgMigrations) Applied(ctx context.Context) (map[string]struct{}, error) {
	rows, err := m.pool.Query(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("query schema_migrations: %w", err)
	}
	defer rows.Close()

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

func (m *pgMigrations) Apply(ctx context.Context, dir, filename string) error {
	content, err := os.ReadFile(filepath.Join(dir, filename))
	if err != nil {
		return fmt.Errorf("read migration %s: %w", filename, err)
	}

	log.Printf("Applying migration %s", filename)
	if len(strings.TrimSpace(string(content))) == 0 {
		return fmt.Errorf("migration %s is empty", filename)
	}

	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", filename, err)
	}

	if _, err := tx.Exec(ctx, string(content)); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("execute migration %s: %w", filename, err)
	}

	if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations(version) VALUES ($1)", filename); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("record migration %s: %w", filename, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit migration %s: %w", filename, err)
	}
	return nil
}

func (m *pgMigrations) Record(ctx context.Context, filename string) error {
	_, err := m.pool.Exec(ctx, "INSERT INTO schema_migrations(version) VALUES ($1) ON CONFLICT DO NOTHING", filename)
	return err
}

func (m *pgMigrations) DropAll(ctx context.Context) error {
	_, err := m.pool.Exec(ctx, `
DO $$
DECLARE
	r RECORD;
BEGIN
	FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP
		EXECUTE 'DROP TABLE IF EXISTS public.' || quote_ident(r.tablename) || ' CASCADE';
	END LOOP;
END $$;
`)
	return err
}
