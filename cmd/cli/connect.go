package main

import (
	"context"
	"fmt"
	"log"
	"net/url"

	"github.com/gracchi-stdio/castogo/internal/config"
	"github.com/gracchi-stdio/castogo/internal/repository/libsql"
	"github.com/gracchi-stdio/castogo/internal/repository/postgres"
	"github.com/urfave/cli/v3"
)

type database interface {
	Ping(ctx context.Context) error
	Close() error
	Migrations() migrationHandler
	// MigrationsDir is the dialect's default migrations directory, used when
	// the --dir flag is empty. The dialect owns this mapping (postgres →
	// sql/migrations, libsql → sql/migrations-sqlite) so the commands never
	// re-parse the URL scheme themselves.
	MigrationsDir() string
}

func connectDatabase(ctx context.Context, cmd *cli.Command) (database, error) {
	databaseURL, err := resolveDatabaseURL(cmd)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	switch u.Scheme {
	case "postgres", "postgresql":
		pool, err := postgres.NewPool(ctx, databaseURL)
		if err != nil {
			return nil, fmt.Errorf("connect postgres: %w", err)
		}
		return &pgDatabase{pool: pool}, nil

	case "libsql", "http", "https":
		db, err := libsql.NewDB(ctx, databaseURL)
		if err != nil {
			return nil, fmt.Errorf("connect libsql: %w", err)
		}
		return &libsqlDatabase{db: db}, nil
	default:
		return nil, fmt.Errorf("unsupported database URL scheme %q - want postgres://.. or https://.. ", u.Scheme)
	}
}

func resolveDatabaseURL(cmd *cli.Command) (string, error) {
	if err := config.LoadConfig(); err != nil {
		return "", err
	}
	if url := cmd.Root().String("database-url"); url != "" {
		config.Cfg.DatabaseURL = url
	}

	return config.Cfg.DatabaseURL, nil
}

func dbPingCommand() *cli.Command {
	return &cli.Command{
		Name:  "db-ping",
		Usage: "Ping database",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			db, err := connectDatabase(ctx, cmd)
			if err != nil {
				return err
			}
			defer db.Close()

			if err := db.Ping(ctx); err != nil {
				return err
			}

			log.Println("database OK!")

			return nil
		},
	}
}
