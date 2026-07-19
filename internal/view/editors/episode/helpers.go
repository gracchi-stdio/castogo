package episodeForm

import (
	"fmt"

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
