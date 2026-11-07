package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gracchi-stdio/castogo/internal/repository"
	"github.com/gracchi-stdio/castogo/internal/sqlcsqlite"
)

var _ repository.AnalyticsRepository = (*AnalyticRepo)(nil)

type AnalyticRepo struct {
	q *sqlcsqlite.Queries
}

func NewAnalyticRepo(db *sql.DB) *AnalyticRepo {
	return &AnalyticRepo{q: sqlcsqlite.New(db)}
}

// AccumulateBytes implements [repository.AnalyticsRepository].
func (a *AnalyticRepo) AccumulateBytes(ctx context.Context, hash string, episodeID int64, date time.Time, bytes int64) (bytesSeen int64, counted bool, err error) {
	acc, err := a.q.UpsertAccumulator(ctx, sqlcsqlite.UpsertAccumulatorParams{
		Hash:      hash,
		EpisodeID: episodeID,
		Date:      dateToSQLite(date),
		BytesSeen: bytes,
	})
	if err != nil {
		return 0, false, fmt.Errorf("upsert accumulator: %w", err)
	}
	return acc.BytesSeen, acc.Counted, nil
}

// InsertProcessedFile implements [repository.AnalyticsRepository].
func (a *AnalyticRepo) InsertProcessedFile(ctx context.Context, filePath string, fileDate time.Time, entriesCount int) error {
	_, err := a.q.InsertProcessedFile(ctx, sqlcsqlite.InsertProcessedFileParams{
		Filename:     filePath,
		FileDate:     dateToSQLite(fileDate),
		EntriesCount: int64(entriesCount),
	})
	if err != nil {
		return fmt.Errorf("insert processed file: %w", err)
	}
	return nil
}

// IsFileProcessed implements [repository.AnalyticsRepository].
func (a *AnalyticRepo) IsFileProcessed(ctx context.Context, filePath string) (bool, error) {
	_, err := a.q.GetProcessedFileByName(ctx, filePath)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return false, fmt.Errorf("get processed file: %w", err)
}

// MarkCounted implements [repository.AnalyticsRepository].
func (a *AnalyticRepo) MarkCounted(ctx context.Context, hash string) error {
	err := a.q.MarkAccumulatorCounted(ctx, hash)
	if err != nil {
		return fmt.Errorf("mark accumulator counted: %w", err)
	}
	return nil
}

// UpsertByCountry implements [repository.AnalyticsRepository].
func (a *AnalyticRepo) UpsertByCountry(ctx context.Context, podcastID int64, date time.Time, countryCode string) error {
	_, err := a.q.UpsertPodcastByCountry(ctx, sqlcsqlite.UpsertPodcastByCountryParams{
		PodcastID:   podcastID,
		Date:        dateToSQLite(date),
		CountryCode: countryCode,
	})
	if err != nil {
		return fmt.Errorf("upsert by country: %w", err)
	}
	return nil
}

// UpsertByEpisode implements [repository.AnalyticsRepository].
func (a *AnalyticRepo) UpsertByEpisode(ctx context.Context, podcastID int64, episodeID int64, date time.Time, age int) error {
	_, err := a.q.UpsertPodcastByEpisode(ctx, sqlcsqlite.UpsertPodcastByEpisodeParams{
		PodcastID: podcastID,
		EpisodeID: episodeID,
		Date:      dateToSQLite(date),
		Age:       int64(age),
	})
	if err != nil {
		return fmt.Errorf("upsert by episode: %w", err)
	}
	return nil
}

// UpsertByHour implements [repository.AnalyticsRepository].
func (a *AnalyticRepo) UpsertByHour(ctx context.Context, podcastID int64, date time.Time, hour int) error {
	_, err := a.q.UpsertPodcastByHour(ctx, sqlcsqlite.UpsertPodcastByHourParams{
		PodcastID: podcastID,
		Date:      dateToSQLite(date),
		Hour:      int64(hour),
	})
	if err != nil {
		return fmt.Errorf("upsert by hour: %w", err)
	}
	return nil
}

// UpsertByPlayer implements [repository.AnalyticsRepository].
func (a *AnalyticRepo) UpsertByPlayer(ctx context.Context, podcastID int64, date time.Time, service string, app string, device string, os string, isBot bool) error {
	_, err := a.q.UpsertPodcastByPlayer(ctx, sqlcsqlite.UpsertPodcastByPlayerParams{
		PodcastID: podcastID,
		Date:      dateToSQLite(date),
		Service:   service,
		App:       app,
		Device:    device,
		Os:        os,
		IsBot:     isBot,
	})
	if err != nil {
		return fmt.Errorf("upsert by player: %w", err)
	}
	return nil
}

// UpsertPodcastDaily implements [repository.AnalyticsRepository].
func (a *AnalyticRepo) UpsertPodcastDaily(ctx context.Context, podcastID int64, date time.Time, bandwidth int64) error {
	_, err := a.q.UpsertPodcastDaily(ctx, sqlcsqlite.UpsertPodcastDailyParams{
		PodcastID: podcastID,
		Date:      dateToSQLite(date),
		Bandwidth: bandwidth,
	})
	if err != nil {
		return fmt.Errorf("upsert podcast daily: %w", err)
	}
	return nil
}
