package episodeForm

import (
	"fmt"
	"net/url"
	"path"
	"time"

	"github.com/gracchi-stdio/castogo/internal/domain"
	selectcomponent "github.com/gracchi-stdio/castogo/internal/view/components/select"
)

// PageLabel is the display label for a page in the companion-page selector.
// Shared by the initial render (pageOptions) and the post-link signal patch in
// the handler (patchPageLink) so the two never drift.
func PageLabel(p *domain.Page) string {
	return p.Title + " (" + p.Path + ")"
}

// pageOptions builds select options for the "link an existing page" dropdown.
func pageOptions(pages []*domain.Page) []selectcomponent.SelectOptionArgs {
	opts := make([]selectcomponent.SelectOptionArgs, 0, len(pages))
	for _, p := range pages {
		opts = append(opts, selectcomponent.SelectOptionArgs{
			Value: fmt.Sprintf("%d", p.ID),
			Label: PageLabel(p),
		})
	}
	return opts
}

// formatDuration turns a duration in seconds into a compact m:ss / h:m string.
// Mirrors the helpers in episodeview and pageview/blocks (per-package convention).
func formatDuration(seconds int) string {
	if seconds <= 0 {
		return "—"
	}
	m := seconds / 60
	s := seconds % 60
	if m >= 60 {
		h := m / 60
		return fmt.Sprintf("%dh %dm", h, m%60)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

// formatFileSize humanizes a byte count (e.g. 12.3 MB). Returns "—" for <= 0.
func formatFileSize(bytes int64) string {
	if bytes <= 0 {
		return "—"
	}
	const unit = 1024
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// audioFilename returns the basename of an AudioSourceURL for display. Falls
// back to the full URL if parsing yields nothing.
func audioFilename(url string) string {
	if url == "" {
		return "—"
	}
	base := path.Base(url)
	if base == "." || base == "/" || base == "" {
		return url
	}
	return base
}

// audioActionLabel returns the verb for the audio card's primary action.
func audioActionLabel(hasAudio bool) string {
	if hasAudio {
		return "Replace Audio"
	}
	return "Add Audio"
}

// audioMetadataHandler is the data-on:audiometadata expression for the create
// form: it copies the decoded preview fields (display-only) into signals and
// flags extraction complete. The server recomputes real metadata from the
// processed audio, so numeric meta_* fields are intentionally not wired.
const audioMetadataHandler = `
$audio_duration = evt.detail.audio_duration;
$audio_sample_rate = evt.detail.audio_sample_rate;
$audio_channel_count = evt.detail.audio_channel_count;
$audio_bitrate = evt.detail.audio_bitrate;
$audio_format = evt.detail.audio_format;
$audio_mime_type = evt.detail.audio_mime_type;
$audio_file_size = evt.detail.audio_file_size;
$metadata_extracted = true;
`

// statusColor maps an episode status to badge Tailwind classes (with dark-mode
// variants). Ported from the former episodeview package.
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

// publishDateValue formats a publish-at time as a YYYY-MM-DD calendar value, or
// "" when unset. Ported from the former episodeview package.
func publishDateValue(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

// episodeListURL builds the admin episode list URL with the given status/search
// query params, omitting empty values so "All + no search" yields a clean path.
func episodeListURL(status, search string) string {
	q := url.Values{}
	if status != "" {
		q.Set("status", status)
	}
	if search != "" {
		q.Set("filter", search)
	}
	if enc := q.Encode(); enc != "" {
		return "/admin/episodes?" + enc
	}
	return "/admin/episodes"
}

// pillClass returns Tailwind classes for a filter pill, highlighting the active
// one with the primary token.
func pillClass(active bool) string {
	if active {
		return "inline-flex items-center rounded-md px-2.5 py-1 text-sm bg-primary text-primary-foreground"
	}
	return "inline-flex items-center rounded-md px-2.5 py-1 text-sm text-muted-foreground hover:bg-accent"
}
