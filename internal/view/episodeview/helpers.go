package episodeview

import (
	"fmt"
	"time"

	"github.com/gracchi-stdio/castogo/internal/domain"
)

func formatDuration(seconds int) string {
	if seconds <= 0 {
		return "—"
	}
	m := seconds / 60
	s := seconds % 60
	if m >= 60 {
		h := m / 60
		m = m % 60
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

func statusColor(status domain.EpisodeStatus) string {
	switch status {
	case domain.EpisodeStatusPublished:
		return "bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400"
	case domain.EpisodeStatusDraft:
		return "bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400"
	case domain.EpisodeStatusScheduled:
		return "bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400"
	case domain.EpisodeStatusArchived:
		return "bg-gray-100 text-gray-500 dark:bg-gray-800 dark:text-gray-400"
	default:
		return "bg-secondary text-secondary-foreground"
	}
}

func publishDateValue(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}
