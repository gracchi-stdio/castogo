package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/urfave/cli/v3"
)

// migrationHandler is the dialect contract for the migration runner. Each
// backend (pgMigrations, sqliteMigrations) implements it; the commands below
// are written against it and never know which database they talk to.
type migrationHandler interface {
	EnsureMigrationsTable(ctx context.Context) error
	Applied(ctx context.Context) (map[string]struct{}, error)
	Apply(ctx context.Context, dir, filename string) error
	Record(ctx context.Context, filename string) error
	DropAll(ctx context.Context) error
	Lock(ctx context.Context) (func(), error)
}

var errNotImplemented = errors.New("not implemented: dialect handler stub")

// listMigrationFiles is dialect-neutral: it only reads the directory.
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

func migrateCommand() *cli.Command {
	return &cli.Command{
		Name:  "migrate",
		Usage: "Apply pending SQL migrations",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "dir",
				Usage: "Migrations directory",
				// Step 2: default becomes empty and resolves by URL scheme —
				// sql/migrations vs sql/migrations-sqlite
				Value: filepath.Join("sql", "migrations"),
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "List pending migrations without applying",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			db, err := connectDatabase(ctx, cmd)
			if err != nil {
				return err
			}
			defer db.Close()

			m := db.Migrations()

			release, err := m.Lock(ctx)
			if err != nil {
				return err
			}
			defer release()

			if err := m.EnsureMigrationsTable(ctx); err != nil {
				return fmt.Errorf("ensure schema_migrations: %w", err)
			}

			dir := cmd.String("dir")
			files, err := listMigrationFiles(dir)
			if err != nil {
				return err
			}

			applied, err := m.Applied(ctx)
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
				if err := m.Apply(ctx, dir, file); err != nil {
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
	}
}

func migrateResetCommand() *cli.Command {
	return &cli.Command{
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

			db, err := connectDatabase(ctx, cmd)
			if err != nil {
				return err
			}
			defer db.Close()

			m := db.Migrations()

			if err := m.DropAll(ctx); err != nil {
				return fmt.Errorf("drop all tables: %w", err)
			}

			if err := m.EnsureMigrationsTable(ctx); err != nil {
				return fmt.Errorf("ensure schema_migrations: %w", err)
			}

			dir := cmd.String("dir")
			files, err := listMigrationFiles(dir)
			if err != nil {
				return err
			}

			for _, file := range files {
				if err := m.Apply(ctx, dir, file); err != nil {
					return err
				}
			}

			log.Printf("Reset complete, applied %d migration(s)", len(files))
			return nil
		},
	}
}

func migrateAdoptCommand() *cli.Command {
	return &cli.Command{
		Name:  "migrate-adopt",
		Usage: "Record migration files as applied WITHOUT running them (for databases whose schema was created another way, e.g. docker-entrypoint-initdb.d or manual psql)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "dir",
				Usage: "Migrations directory",
				Value: filepath.Join("sql", "migrations"),
			},
			&cli.BoolFlag{
				Name: "yes",
				Usage: "Confirm adopt (skipped migrations are never applied later)",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if !cmd.Bool("yes") {
				return fmt.Errorf("refusing to adopt without --yes")
			}

			db, err := connectDatabase(ctx, cmd)
			if err != nil {
				return err
			}
			defer db.Close()

			m := db.Migrations()

			release, err := m.Lock(ctx)
			if err != nil {
				return err
			}
			defer release()

			if err := m.EnsureMigrationsTable(ctx); err != nil {
				return fmt.Errorf("ensure schema_migrations: %w", err)
			}

			files, err := listMigrationFiles(cmd.String("dir"))
			if err != nil {
				return err
			}

			applied, err := m.Applied(ctx)
			if err != nil {
				return err
			}

			adopted := 0
			for _, file := range files {
				if _, ok := applied[file]; ok {
					continue
				}
				if err := m.Record(ctx, file); err != nil {
					return fmt.Errorf("adopt migration %s: %w", file, err)
				}
				log.Printf("Adopted %s", file)
				adopted++
			}

			if adopted == 0 {
				log.Println("Nothing to adopt — all migration files already recorded")
			}
			return nil
		},
	}
}
