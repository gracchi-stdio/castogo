package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"

	"github.com/gracchi-stdio/castogo/internal/repository"
	"github.com/gracchi-stdio/castogo/internal/repository/postgres"
	"github.com/gracchi-stdio/castogo/internal/repository/sqlite"
	"github.com/gracchi-stdio/castogo/internal/session"
)

type database struct {
	ping          func(ctx context.Context) error
	close         func()
	sessionStore  session.SessionStore
	userRepo      repository.UserRepository
	episodeRepo   repository.EpisodeRepository
	podcastRepo   repository.PodcastConfigRepository
	pageRepo      repository.PageRepository
	analyticsRepo repository.AnalyticsRepository
}

func connectDatabase(ctx context.Context, databaseURL, sessionSecret string) (*database, error) {
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
		store, err := session.NewPGStore(pool, []byte(sessionSecret))
		if err != nil {
			pool.Close()
			return nil, fmt.Errorf("create session store: %w", err)
		}
		return &database{
			ping:          pool.Ping,
			close:         pool.Close,
			sessionStore:  store,
			userRepo:      postgres.NewUserRepo(pool),
			episodeRepo:   postgres.NewEpisodeRepo(pool),
			podcastRepo:   postgres.NewPodcastConfigRepository(pool),
			pageRepo:      postgres.NewPageRepo(pool),
			analyticsRepo: postgres.NewAnalyticsRepo(pool),
		}, nil

	case "libsql", "http", "https":
		db, err := sqlite.NewRemoteDB(ctx, databaseURL)
		if err != nil {
			return nil, fmt.Errorf("connect libsql: %w", err)
		}
		return connectSQLiteDatabase(db, sessionSecret)

	case "file", "sqlite":
		db, err := sqlite.NewLocalDB(ctx, databaseURL)
		if err != nil {
			return nil, fmt.Errorf("connect sqlite: %w", err)
		}
		return connectSQLiteDatabase(db, sessionSecret)

	default:
		return nil, fmt.Errorf("unsupported database URL scheme %q - want postgres://, libsql:// or file:", u.Scheme)

	}

}

func connectSQLiteDatabase(db *sql.DB, sessionSecret string) (*database, error) {
	store, err := session.NewSQLiteStore(db, []byte(sessionSecret))
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create session store: %w", err)
	}
	return &database{
		ping:          db.PingContext,
		close:         func() { _ = db.Close() },
		sessionStore:  store,
		userRepo:      sqlite.NewUserRepo(db),
		episodeRepo:   sqlite.NewEpisodeRepo(db),
		podcastRepo:   sqlite.NewPodcastConfigRepo(db),
		pageRepo:      sqlite.NewPageRepo(db),
		analyticsRepo: sqlite.NewAnalyticRepo(db),
	}, nil
}
