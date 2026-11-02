package postgres

import (
	"context"
	"time"

	"github.com/gracchi-stdio/castogo/internal/sqlcpg"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AnalyticsPostgres struct {
	q *sqlcpg.Queries
}

func NewAnalyticsPostgres(pool *pgxpool.Pool) *AnalyticsPostgres {
	return &AnalyticsPostgres{
		q: sqlcpg.New(pool),
	}
}

func (r *AnalyticsPostgres) InsertProcessedFile(ctx context.Context, filePath string, fileDate time.Time, entriesCount int) error {
	_, err := r.q.InsertProcessedFile(ctx, sqlcpg.InsertProcessedFileParams{
		Filename:     filePath,
		FileDate:     pgtype.Date{Time: fileDate, Valid: true},
		EntriesCount: int32(entriesCount),
	})
	return err
}

func (r *AnalyticsPostgres) IsFileProcessed(ctx context.Context, filePath string) (bool, error) {
	_, err := r.q.GetProcessedFileByName(ctx, filePath)
	if err == nil {
		return true, nil
	}
	if err == pgx.ErrNoRows {
		return false, nil
	}
	return false, err
}

func (r *AnalyticsPostgres) AccumulateBytes(ctx context.Context, hash string, episodeID int64, date time.Time, bytes int64) (int64, bool, error) {
	acc, err := r.q.UpsertAccumulator(ctx, sqlcpg.UpsertAccumulatorParams{
		Hash:      hash,
		EpisodeID: episodeID,
		Date:      pgtype.Date{Time: date, Valid: true},
		BytesSeen: bytes,
	})
	if err != nil {
		return 0, false, err
	}
	return acc.BytesSeen, acc.Counted, err
}

func (r *AnalyticsPostgres) MarkCounted(ctx context.Context, hash string) error {
	return r.q.MarkAccumulatorCounted(ctx, hash)
}

func (r *AnalyticsPostgres) UpsertPodcastDaily(ctx context.Context, podcastID int64, date time.Time, bandwidth int64) error {
	_, err := r.q.UpsertPodcastDaily(ctx, sqlcpg.UpsertPodcastDailyParams{
		PodcastID: podcastID,
		Date:      pgtype.Date{Time: date, Valid: true},
		Bandwidth: bandwidth,
	})
	return err
}

func (r *AnalyticsPostgres) UpsertByEpisode(ctx context.Context, podcastID, episodeID int64, date time.Time, age int) error {
	_, err := r.q.UpsertPodcastByEpisode(ctx, sqlcpg.UpsertPodcastByEpisodeParams{
		PodcastID: podcastID,
		EpisodeID: episodeID,
		Date:      pgtype.Date{Time: date, Valid: true},
		Age:       int32(age),
	})
	return err
}

func (r *AnalyticsPostgres) UpsertByPlayer(ctx context.Context, podcastID int64, date time.Time, service, app, device, os string, isBot bool) error {
	_, err := r.q.UpsertPodcastByPlayer(ctx, sqlcpg.UpsertPodcastByPlayerParams{
		PodcastID: podcastID,
		Date:      pgtype.Date{Time: date, Valid: true},
		Service:   service,
		App:       app,
		Device:    device,
		Os:        os,
		IsBot:     isBot,
	})
	return err
}

func (r *AnalyticsPostgres) UpsertByHour(ctx context.Context, podcastID int64, date time.Time, hour int) error {
	_, err := r.q.UpsertPodcastByHour(ctx, sqlcpg.UpsertPodcastByHourParams{
		PodcastID: podcastID,
		Date:      pgtype.Date{Time: date, Valid: true},
		Hour:      int32(hour),
	})
	return err
}

func (r *AnalyticsPostgres) UpsertByCountry(ctx context.Context, podcastID int64, date time.Time, countryCode string) error {
	_, err := r.q.UpsertPodcastByCountry(ctx, sqlcpg.UpsertPodcastByCountryParams{
		PodcastID:   podcastID,
		Date:        pgtype.Date{Time: date, Valid: true},
		CountryCode: countryCode,
	})
	return err
}
