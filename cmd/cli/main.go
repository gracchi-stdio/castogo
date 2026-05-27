package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gracchi-stdio/castogo/internal/config"
	"github.com/gracchi-stdio/castogo/internal/repository/postgres"
	"github.com/gracchi-stdio/castogo/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "castogo-cli",
		Usage: "CLI utilities for Castogo",
		Commands: []*cli.Command{
			{
				Name:  "register",
				Usage: "Register a new user",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "email",
						Usage:    "Email for registration",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "password",
						Usage:    "Password for registration",
						Required: true,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					db, err := connectDB(ctx)
					if err != nil {
						return err
					}
					defer db.Close()

					userRepo := postgres.NewUserRepo(db)
					authService := service.NewAuthService(userRepo)

					email := cmd.String("email")
					password := cmd.String("password")

					user, err := authService.Register(ctx, email, password)
					if err != nil {
						return fmt.Errorf("failed to create user: %w", err)
					}
					log.Printf("user %s successfully created", user.Email)
					return nil
				},
			},
			{
				Name:  "migrate-reset",
				Usage: "Drop all tables and re-apply SQL migrations",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "dir",
						Usage: "Migrations directory",
						Value: filepath.Join("sql", "migrations"),
					},
					&cli.BoolFlag{
						Name:  "yes",
						Usage: "Confirm destructive reset",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if !cmd.Bool("yes") {
						return fmt.Errorf("refusing to reset without --yes")
					}

					db, err := connectDB(ctx)
					if err != nil {
						return err
					}
					defer db.Close()

					if err := dropAllTables(ctx, db); err != nil {
						return fmt.Errorf("drop all tables: %w", err)
					}

					if err := ensureSchemaMigrations(ctx, db); err != nil {
						return fmt.Errorf("ensure schema_migrations: %w", err)
					}

					migrationsDir := cmd.String("dir")
					files, err := listMigrationFiles(migrationsDir)
					if err != nil {
						return err
					}

					for _, file := range files {
						if err := applyMigration(ctx, db, migrationsDir, file); err != nil {
							return err
						}
					}

					log.Printf("Reset complete, applied %d migration(s)", len(files))
					return nil
				},
			},
			{
				Name:  "migrate",
				Usage: "Apply pending SQL migrations",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "dir",
						Usage: "Migrations directory",
						Value: filepath.Join("sql", "migrations"),
					},
					&cli.BoolFlag{
						Name:  "dry-run",
						Usage: "List pending migrations without applying",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					db, err := connectDB(ctx)
					if err != nil {
						return err
					}
					defer db.Close()

					if err := ensureSchemaMigrations(ctx, db); err != nil {
						return fmt.Errorf("ensure schema_migrations: %w", err)
					}

					migrationsDir := cmd.String("dir")
					files, err := listMigrationFiles(migrationsDir)
					if err != nil {
						return err
					}

					applied, err := loadAppliedMigrations(ctx, db)
					if err != nil {
						return err
					}

					pending := make([]string, 0)
					for _, file := range files {
						if _, ok := applied[file]; ok {
							continue
						}
						pending = append(pending, file)
					}

					if cmd.Bool("dry-run") {
						if len(pending) == 0 {
							log.Println("No pending migrations")
							return nil
						}
						log.Println("Pending migrations:")
						for _, file := range pending {
							log.Printf("- %s", file)
						}
						return nil
					}

					for _, file := range pending {
						if err := applyMigration(ctx, db, migrationsDir, file); err != nil {
							return err
						}
					}

					if len(pending) == 0 {
						log.Println("No pending migrations")
					} else {
						log.Printf("Applied %d migration(s)", len(pending))
					}
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func connectDB(ctx context.Context) (*pgxpool.Pool, error) {
	if err := config.LoadConfig(); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	db, err := postgres.NewPool(ctx, config.Cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection successful")
	return db, nil
}

func ensureSchemaMigrations(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version TEXT PRIMARY KEY,
	applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`)
	return err
}

func listMigrationFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".sql") {
			files = append(files, name)
		}
	}

	sort.Strings(files)
	return files, nil
}

func loadAppliedMigrations(ctx context.Context, db *pgxpool.Pool) (map[string]struct{}, error) {
	rows, err := db.Query(ctx, "SELECT version FROM schema_migrations")
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

func applyMigration(ctx context.Context, db *pgxpool.Pool, dir, filename string) error {
	path := filepath.Join(dir, filename)
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", filename, err)
	}

	log.Printf("Applying migration %s", filename)
	if len(strings.TrimSpace(string(content))) == 0 {
		return fmt.Errorf("migration %s is empty", filename)
	}

	tx, err := db.Begin(ctx)
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

func dropAllTables(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
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
