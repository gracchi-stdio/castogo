package service

import (
	"context"

	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/gracchi-stdio/castogo/internal/repository"
)

// Episode numbers are DERIVED from publication order, not stored. This file is
// the single place that computes them, shared by EpisodeService (admin list,
// showcase, search) and FeedService (RSS <itunes:episode>).

// episodeNumberRanks returns a map of episode ID -> 1-based chronological rank,
// built from the repo's ranking list (all non-archived episodes with a
// publish_at, in publish_at/id order). Drafts and archived episodes are absent.
func episodeNumberRanks(ctx context.Context, repo repository.EpisodeRepository) (map[int64]int, error) {
	ranked, err := repo.ListForRanking(ctx)
	if err != nil {
		return nil, err
	}
	ranks := make(map[int64]int, len(ranked))
	for i, ep := range ranked {
		ranks[ep.ID] = i + 1
	}
	return ranks, nil
}

// applyEpisodeNumbers sets each episode's EpisodeNumber from the rank map.
// Episodes absent from the map (drafts, archived) keep number 0.
func applyEpisodeNumbers(eps []*domain.Episode, ranks map[int64]int) {
	for _, ep := range eps {
		ep.EpisodeNumber = ranks[ep.ID]
	}
}
