package repository

import (
	"context"
	"time"
)

type AnalyticsRepository interface {
	// pipeline states for processing analytics files
	InsertProcessedFile(ctx context.Context, filePath string, fileDate time.Time, entriesCount int) error
	IsFileProcessed(ctx context.Context, filePath string) (bool, error)

	// IAB accumulator
	AccumulateBytes(ctx context.Context, hash string, episodeID int64, date time.Time, bytes int64) (bytesSeen int64, counted bool, err error)
	MarkCounted(ctx context.Context, hash string) error

	// Summary upserts
	UpsertPodcastDaily(ctx context.Context, podcastID int64, date time.Time, bandwidth int64) error
	UpsertByHour(ctx context.Context, podcastID int64, date time.Time, hour int) error
	UpsertByEpisode(ctx context.Context, podcastID, episodeID int64, date time.Time, age int) error
	UpsertByPlayer(ctx context.Context, podcastID int64, date time.Time, service, app, device, os string, isBot bool) error
	UpsertByCountry(ctx context.Context, podcastID int64, date time.Time, countryCode string) error
}
