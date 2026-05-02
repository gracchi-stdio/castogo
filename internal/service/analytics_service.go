package service

import (
	"context"
	"crypto/sha1"
	"fmt"
	"time"

	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/gracchi-stdio/castogo/internal/repository"
)

type AnalyticsService struct {
	logFetcher LogFetcher
	repo       repository.AnalyticsRepository
	episodes   repository.EpisodeRepository
}

func NewAnalyticsService(
	fetcher LogFetcher,
	repo repository.AnalyticsRepository,
	episodes repository.EpisodeRepository,
) *AnalyticsService {
	return &AnalyticsService{
		logFetcher: fetcher,
		repo:       repo,
		episodes:   episodes,
	}
}

func (s *AnalyticsService) ProcessLogsForDate(ctx context.Context, date time.Time) error {
	entries, err := s.logFetcher.FetchEntries(ctx, date)
	if err != nil {
		return fmt.Errorf("fetching log entries: %w", err)
	}

	fmt.Printf("%s[analytics]%s fetched %d entries for %s\n", cCyan, cReset, len(entries), date.Format("2006-01-02"))

	publishedEpisodes, err := s.episodes.ListPublished(ctx, 500, 0)
	if err != nil {
		return fmt.Errorf("fetching published episodes: %w", err)
	}

	episodeMap := make(map[string]domain.EpisodeMetadata)
	for _, ep := range publishedEpisodes {
		episodeMap[ep.AudioSourceURL] = domain.EpisodeMetadata{
			EpisodeID:   ep.ID,
			PodcastID:   1, // single podcast for now, can be extended later
			FileSize:    ep.AudioMetadata.FileSize,
			Duration:    int64(ep.Duration),
			PublishedAt: ep.CreatedAt,
		}
	}

	filtered := 0
	downloadCount := 0

	for _, rawEntry := range entries {
		if !shouldProcess(rawEntry) {
			filtered++
			continue
		}

		meta, ok := episodeMap[rawEntry.URL]
		if !ok {
			filtered++
			continue // URL not in our published episodes, skip
		}

		h := sha1.Sum([]byte(fmt.Sprintf("%s-%s-%s-%d", date.Format("2006-01-02"), rawEntry.ClientIP, rawEntry.UserAgent, meta.EpisodeID)))
		hash := fmt.Sprintf("%x", h[:])

		accBytes, counted, err := s.repo.AccumulateBytes(ctx, hash, meta.EpisodeID, date, rawEntry.BytesSent)
		if err != nil {
			return fmt.Errorf("accumulating bytes: %w", err)
		}

		if meta.Duration == 0 {
			continue // can't calculate threshold without duration
		}
		threshold := int64(float64(meta.FileSize) * (60.0 / float64(meta.Duration)))
		if !counted && accBytes >= threshold {
			downloadCount++
			if err := s.repo.MarkCounted(ctx, hash); err != nil {
				return fmt.Errorf("marking counted: %w", err)
			}

			// Upsert all 6 summary tables
			if err := s.repo.UpsertPodcastDaily(ctx, meta.PodcastID, date, rawEntry.BytesSent); err != nil {
				return fmt.Errorf("upserting podcast daily: %w", err)
			}

			age := int(date.Sub(meta.PublishedAt).Hours() / 24)
			if err := s.repo.UpsertByEpisode(ctx, meta.PodcastID, meta.EpisodeID, date, age); err != nil {
				return fmt.Errorf("upserting by episode: %w", err)
			}

			hour := time.UnixMilli(rawEntry.Timestamp).Hour()
			if err := s.repo.UpsertByHour(ctx, meta.PodcastID, date, hour); err != nil {
				return fmt.Errorf("upserting by hour: %w", err)
			}

			ua := ParseUserAgent(rawEntry.UserAgent)
			if err := s.repo.UpsertByPlayer(ctx, meta.PodcastID, date, ua.Service, ua.App, ua.Device, ua.OS, ua.IsBot); err != nil {
				return fmt.Errorf("upserting by player: %w", err)
			}

			if err := s.repo.UpsertByCountry(ctx, meta.PodcastID, date, rawEntry.CountryCode); err != nil {
				return fmt.Errorf("upserting by country: %w", err)
			}
		}
	}

	fmt.Printf("%s[analytics]%s %s summary: %d fetched, %d filtered, %d new downloads counted\n",
		cCyan, cReset, date.Format("2006-01-02"), len(entries), filtered, downloadCount)

	return nil
}

func shouldProcess(entry domain.RawLogEntry) bool {
	return entry.StatusCode == 200 || entry.StatusCode == 206
}
