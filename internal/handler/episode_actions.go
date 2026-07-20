package handler

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/gracchi-stdio/castogo/internal/service"
	episodeForm "github.com/gracchi-stdio/castogo/internal/view/editors/episode"
	"github.com/gracchi-stdio/castogo/internal/view/episodeview"
	"github.com/labstack/echo/v5"
	"github.com/starfederation/datastar-go/datastar"
)

// episodeCreateAction handles POST /admin/episodes/create — multipart (audio).
// The decode → validate → process/upload → create flow follows the standard
// recipe; the audio pipeline is extracted into processAndUploadAudio.
func (h *AdminHandler) episodeCreateAction(c *echo.Context) error {
	if err := c.Request().ParseMultipartForm(100 << 20); err != nil {
		patchSignals(c, map[string]string{"uploading": ""})
		return toast(c, "Failed to parse form", "error")
	}

	input := &episodeCreateInput{
		Title:       c.FormValue("title"),
		Description: c.FormValue("description"),
	}
	if err := validate.Struct(input); err != nil {
		errs := fieldValidationErrors(err, input)
		errs["uploading"] = ""
		return patchSignals(c, errs)
	}

	file, header, err := c.Request().FormFile("audio_file")
	if err != nil {
		return patchSignals(c, map[string]string{"audio_file_error": "Audio file is required", "uploading": ""})
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !isAllowedAudioExt(ext) {
		return patchSignals(c, map[string]string{"audio_file_error": "Unsupported file type. Use mp3, wav, m4a, ogg, or flac.", "uploading": ""})
	}

	patchSignals(c, map[string]string{"uploading_status": "Processing audio..."})
	epSlug := slug.Make(input.Title)
	cdnURL, audioMeta, err := h.processAndUploadAudio(c.Request().Context(), file, ext, epSlug)
	if err != nil {
		log.Printf("episode audio processing failed: %v", err)
		patchSignals(c, map[string]string{"uploading": ""})
		return toast(c, "Failed to process audio file", "error")
	}

	patchSignals(c, map[string]string{"uploading_status": "Creating episode record..."})
	episode := &domain.Episode{
		Title:          input.Title,
		Slug:           epSlug,
		Description:    input.Description,
		Duration:       audioMeta.Duration,
		AudioMetadata:  audioMeta,
		AudioSourceURL: cdnURL,
	}
	if _, err := h.episodeService.Create(c.Request().Context(), episode); err != nil {
		log.Printf("create episode failed: %v", err)
		patchSignals(c, map[string]string{"uploading": ""})
		return toast(c, "Failed to create episode", "error")
	}

	log.Printf("Episode created: title=%s audio=%s", input.Title, cdnURL)
	patchSignals(c, map[string]string{"uploading": ""})
	return navigate(c, "/admin/episodes", "", "")
}

// processAndUploadAudio normalizes and uploads the episode audio, returning the
// CDN URL and server-derived metadata. Temp files are cleaned up before return.
func (h *AdminHandler) processAndUploadAudio(ctx context.Context, file io.Reader, ext, slug string) (string, domain.AudioMetadata, error) {
	tmp, err := service.NewTempFile(file, ext)
	if err != nil {
		return "", domain.AudioMetadata{}, err
	}
	defer tmp.Cleanup()

	processResult, err := h.audioProcessor.Process(ctx, tmp.Path, service.DefaultProcessingOptions())
	if err != nil {
		return "", domain.AudioMetadata{}, err
	}
	defer os.Remove(processResult.OutputPath)

	processedFile, err := os.Open(processResult.OutputPath)
	if err != nil {
		return "", domain.AudioMetadata{}, err
	}
	defer processedFile.Close()

	cdnURL, err := h.storageService.UploadFile(ctx, processedFile, episodeAudioFilename(slug))
	if err != nil {
		return "", domain.AudioMetadata{}, err
	}

	meta := domain.AudioMetadata{
		Duration:     int(processResult.Duration),
		SampleRate:   service.DefaultProcessingOptions().TargetSample,
		ChannelCount: 2,
		BitRate:      processResult.Bitrate,
		FileSize:     processResult.FileSize,
		Format:       "mp3",
		MimeType:     "audio/mpeg",
	}
	return cdnURL, meta, nil
}

// episodeAudioFilename builds a unique CDN path: episodes/<slug>-<8hex>.mp3.
func episodeAudioFilename(slug string) string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("episodes/%s-%x.mp3", slug, b)
}

func isAllowedAudioExt(ext string) bool {
	switch ext {
	case ".mp3", ".wav", ".m4a", ".ogg", ".flac":
		return true
	}
	return false
}

// episodeUpdateAction handles POST /admin/episodes/:id — metadata only.
// Publish date, episode audio, and page linkage each have dedicated endpoints.
func (h *AdminHandler) episodeUpdateAction(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid episode ID")
	}

	var raw episodeUpdateInput
	if err := readSignals(c, &raw); err != nil {
		return toast(c, "Invalid request", "error")
	}
	if err := validate.Struct(raw); err != nil {
		return patchFieldErrors(c, err, raw)
	}

	title := raw.Title
	slugVal := raw.Slug
	description := raw.Description
	episodeNumber := raw.EpisodeNumber
	explicit := raw.Explicit.Checked

	update := &domain.UpdateEpisode{
		ID:            id,
		Title:         &title,
		Slug:          &slugVal,
		Description:   &description,
		EpisodeNumber: &episodeNumber,
		Explicit:      &explicit,
	}
	if raw.PublishAt != "" {
		t, err := time.Parse("2006-01-02", raw.PublishAt)
		if err != nil {
			return patchSignals(c, map[string]string{"publish_at_error": "Invalid date"})
		}
		update.PublishAt = &t
	}

	if _, err := h.episodeService.Update(c.Request().Context(), update); err != nil {
		return toast(c, "Failed to update episode", "error")
	}
	return toast(c, "Episode saved successfully", "success")
}

// episodeUpdatePublishAt handles PATCH /admin/episodes/:id/publish-at — the
// quick set-publish-date action from the list row's calendar sheet. Reads the
// calendar's namespaced signal (publish-cal-{id} → publish_cal_{id}.dateValue)
// and patches the row in place.
func (h *AdminHandler) episodeUpdatePublishAt(c *echo.Context) error {
	id := parseInt64(c.Param("id"))
	signalKey := fmt.Sprintf("publish_cal_%d", id)

	var raw map[string]any
	if err := readSignals(c, &raw); err != nil {
		return toast(c, "Invalid input", "error")
	}

	var dateValue string
	if calSignals, ok := raw[signalKey].(map[string]any); ok {
		if dv, ok := calSignals["dateValue"].(string); ok {
			dateValue = dv
		}
	}

	var publishAt *time.Time
	if dateValue != "" {
		t, err := time.Parse("2006-01-02", dateValue)
		if err != nil {
			return toast(c, "Invalid date format", "error")
		}
		publishAt = &t
	}

	updated, err := h.episodeService.Update(c.Request().Context(), &domain.UpdateEpisode{
		ID:        id,
		PublishAt: publishAt,
	})
	if err != nil {
		log.Printf("update publish_at failed: %v", err)
		return toast(c, "Failed to update publish date", "error")
	}

	sse(c).PatchElementTempl(episodeview.EpisodeRow(updated),
		datastar.WithSelectorID(fmt.Sprintf("episode-row-%d", id)),
		datastar.WithModeOuter())
	return nil
}

func (h *AdminHandler) episodeDelete(c *echo.Context) error {
	id := parseInt64(c.Param("id"))

	if err := h.episodeService.Delete(c.Request().Context(), id); err != nil {
		log.Printf("delete episode failed: %v", err)
		return toast(c, "Failed to delete episode", "error")
	}

	sse(c).RemoveElementByID(fmt.Sprintf("episode-row-%d", id))
	return nil
}

// episodeLinkPage links an existing page to the episode, then patches the
// companion-page card in place.
func (h *AdminHandler) episodeLinkPage(c *echo.Context) error {
	id := parseInt64(c.Param("id"))

	var raw episodeLinkPageInput
	if err := readSignals(c, &raw); err != nil {
		return toast(c, "Invalid request", "error")
	}
	if raw.ExistingPageID <= 0 {
		return toast(c, "Choose a page to link", "error")
	}
	if err := h.episodeService.LinkPage(c.Request().Context(), id, raw.ExistingPageID); err != nil {
		return toast(c, "Failed to link page", "error")
	}
	return h.patchPageLink(c, id)
}

// episodeCreateCompanion creates a new page (from the dialog's title/slug) and
// links the episode to it, then patches the companion-page card in place.
func (h *AdminHandler) episodeCreateCompanion(c *echo.Context) error {
	id := parseInt64(c.Param("id"))

	var raw episodeCreateCompanionInput
	if err := readSignals(c, &raw); err != nil {
		return toast(c, "Invalid request", "error")
	}
	if err := validate.Struct(raw); err != nil {
		return patchFieldErrors(c, err, raw)
	}

	if _, err := h.pageService.CreateCompanionPage(c.Request().Context(), id, raw.Title, raw.Slug); err != nil {
		if errors.Is(err, domain.ErrDuplicatePath) {
			return patchSignals(c, map[string]string{"companion_slug_error": "A page with this slug already exists"})
		}
		if errors.Is(err, domain.ErrReservedSlug) {
			return patchSignals(c, map[string]string{"companion_slug_error": "This slug is reserved and cannot be used"})
		}
		return toast(c, "Failed to create companion page", "error")
	}

	// Close the dialog, clear its fields, confirm, bust the cached pages list
	// (a new page now exists), then refresh the card to linked state. The toast
	// fires before the refresh so it shows even if the reload happens to fail
	// (the page was already created at this point).
	sse(c).PatchSignals([]byte(fmt.Sprintf(
		`{"companion_dialog_%d":{"open":false},"companion_title":"","companion_slug":"","companion_title_error":"","companion_slug_error":""}`,
		id,
	)))
	toast(c, "Companion page created", "success")
	bustPagesCache(c, "")
	return h.patchPageLink(c, id)
}

// episodeUnlinkPage removes the episode↔page link, then patches the
// companion-page card back to its unlinked state.
func (h *AdminHandler) episodeUnlinkPage(c *echo.Context) error {
	id := parseInt64(c.Param("id"))

	if err := h.episodeService.UnlinkPage(c.Request().Context(), id); err != nil {
		return toast(c, "Failed to unlink page", "error")
	}
	return h.patchPageLink(c, id)
}

// patchPageLink reloads the episode + pages, rebuilds the companion-page view
// model, and swaps the #episode-page-link-{id} fragment in place over SSE.
func (h *AdminHandler) patchPageLink(c *echo.Context, id int64) error {
	episode, err := h.episodeService.GetByID(c.Request().Context(), id)
	if err != nil {
		return toast(c, "Failed to refresh", "error")
	}
	pages, err := h.pageService.ListPages(c.Request().Context())
	if err != nil {
		return toast(c, "Failed to refresh", "error")
	}

	args := episodeForm.Args{Episode: episode, Pages: pages}
	if episode.LinkedPageID != nil {
		for _, p := range pages {
			if p.ID == *episode.LinkedPageID {
				args.LinkedPage = p
				break
			}
		}
	}

	sse(c).PatchElementTempl(episodeForm.PageLinkContent(args),
		datastar.WithSelectorID(fmt.Sprintf("episode-page-link-%d", id)),
		datastar.WithModeOuter())
	// The re-render's data-signals only declares initial state, so the select's
	// internal value/label would otherwise stay stale — push them explicitly.
	if args.LinkedPage != nil {
		label := episodeForm.PageLabel(args.LinkedPage)
		sse(c).PatchSignals([]byte(fmt.Sprintf(`{"page_select":{"value":%q,"label":%q}}`,
			strconv.FormatInt(args.LinkedPage.ID, 10), label)))
	} else {
		sse(c).PatchSignals([]byte(`{"page_select":{"value":"","label":""}}`))
	}
	// The card patch updates the live page, but Swup still holds the pre-patch
	// snapshot of this edit URL — bust it so back-navigation shows the fresh state.
	bustCache(c, fmt.Sprintf("/admin/episodes/%d/edit", id))
	return nil
}
