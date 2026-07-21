package episodeForm

import (
	"fmt"
	"path"

	"github.com/gracchi-stdio/castogo/internal/domain"
	selectcomponent "github.com/gracchi-stdio/castogo/internal/view/components/select"
)

// PageLabel is the display label for a page in the companion-page selector.
// Shared by the initial render (pageOptions) and the post-link signal patch in
// the handler (patchPageLink) so the two never drift.
func PageLabel(p *domain.Page) string {
	return p.Title + " (/" + p.Path + ")"
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
