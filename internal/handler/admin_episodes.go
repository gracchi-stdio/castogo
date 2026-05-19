package handler

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/gosimple/slug"
	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/gracchi-stdio/castogo/internal/repository"
	"github.com/gracchi-stdio/castogo/internal/service"
	"github.com/gracchi-stdio/castogo/internal/view/episodeview"
	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"
)

func (h *AdminHandler) episodesList(c echo.Context) error {
	searchString := c.QueryParam("filter")
	offset := 0
	if offsetParam := c.QueryParam("offset"); offsetParam != "" {
		offset = parseInt(offsetParam)
	}

	episodes, err := h.episodeService.List(c.Request().Context(), repository.EpisodeFilter{
		Search: searchString,
		Limit:  100,
		Offset: offset,
	})
	if err != nil {
		return echo.NewHTTPError(500, "Failed to load episodes")
	}
	return echo.WrapHandler(
		templ.Handler(episodeview.EpisodesListPage(getSharedData(c), episodes)))(c)
}

func (h *AdminHandler) episodeCreatePage(c echo.Context) error {
	return echo.WrapHandler(templ.Handler(episodeview.EpisodeNewPage(getSharedData(c))))(c)
}

type EpisodeInput struct {
	Title       string `validate:"required"`
	Description string `validate:""`
}

func (h *AdminHandler) episodeCreateAction(c echo.Context) error {
	if err := c.Request().ParseMultipartForm(100 << 20); err != nil {
		sse(c).MarshalAndPatchSignals(map[string]string{"error": "Failed to parse form", "uploading": ""})
		return nil
	}

	input := &EpisodeInput{
		Title:       c.FormValue("title"),
		Description: c.FormValue("description"),
	}

	if err := validate.Struct(input); err != nil {
		errors := fieldValidationErrors(err)
		errors["uploading"] = ""
		sse(c).MarshalAndPatchSignals(errors)
		return nil
	}

	file, header, err := c.Request().FormFile("audio_file")
	if err != nil {
		sse(c).MarshalAndPatchSignals(map[string]string{"audio_file_error": "Audio file is required", "uploading": ""})
		return nil
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".mp3" && ext != ".wav" && ext != ".m4a" && ext != ".ogg" && ext != ".flac" {
		sse(c).MarshalAndPatchSignals(map[string]string{"audio_file_error": "Unsupported file type. Use mp3, wav, m4a, ogg, or flac.", "uploading": ""})
		return nil
	}

	out := sse(c)

	out.MarshalAndPatchSignals(map[string]string{"uploading_status": "Validating file..."})

	tmp, err := service.NewTempFile(file, ext)
	if err != nil {
		sse(c).MarshalAndPatchSignals(map[string]string{"error": "Failed to create temporary file", "uploading": ""})
		return nil
	}
	defer tmp.Cleanup()

	out.MarshalAndPatchSignals(map[string]string{"uploading_status": "Processing audio..."})

	processResult, err := h.audioProcessor.Process(c.Request().Context(), tmp.Path, service.DefaultProcessingOptions())
	if err != nil {
		log.Printf("audio processing failed: %v", err)
		sse(c).MarshalAndPatchSignals(map[string]string{"error": "Failed to process audio file", "uploading": ""})
		return nil
	}
	defer os.Remove(processResult.OutputPath)

	out.MarshalAndPatchSignals(map[string]string{"uploading_status": "Uploading audio..."})

	slug := slug.Make(input.Title)
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	uniqueID := fmt.Sprintf("%x", b)
	filename := fmt.Sprintf("episodes/%s-%s.mp3", slug, uniqueID[:8])

	processedFile, err := os.Open(processResult.OutputPath)
	if err != nil {
		sse(c).MarshalAndPatchSignals(map[string]string{"error": "Failed to open processed audio file", "uploading": ""})
		return nil
	}
	defer processedFile.Close()

	cdnURL, err := h.storageService.UploadFile(c.Request().Context(), processedFile, filename)
	if err != nil {
		log.Printf("upload failed: %v", err)
		out.MarshalAndPatchSignals(map[string]string{"error": "Failed to upload audio file", "uploading": ""})
		return nil
	}

	out.MarshalAndPatchSignals(map[string]string{"uploading_status": "Creating episode record..."})

	audioMetadata := domain.AudioMetadata{
		Duration:     int(processResult.Duration),
		SampleRate:   service.DefaultProcessingOptions().TargetSample,
		ChannelCount: 2,
		BitRate:      processResult.Bitrate,
		FileSize:     processResult.FileSize,
		Format:       "mp3",
		MimeType:     "audio/mpeg",
	}

	episode := &domain.Episode{
		Title:          input.Title,
		Slug:           slug,
		Description:    input.Description,
		Duration:       audioMetadata.Duration,
		AudioMetadata:  audioMetadata,
		AudioSourceURL: cdnURL,
	}

	_, err = h.episodeService.Create(c.Request().Context(), episode)
	if err != nil {
		log.Printf("create episode failed: %v", err)
		out.MarshalAndPatchSignals(map[string]string{"error": "Failed to create episode", "uploading": ""})
		return nil
	}

	out.MarshalAndPatchSignals(map[string]string{"uploading_status": "Finishing up..."})
	time.Sleep(500 * time.Millisecond)

	log.Printf("Episode created: title=%s audio=%s", input.Title, cdnURL)

	out.MarshalAndPatchSignals(map[string]string{"uploading": ""})
	out.ExecuteScript("window.navigateAdmin('/admin/episodes')")
	return nil
}

// updatePublishAtSignals matches the calendar's namespaced signal structure.
// Calendar ID is "publish-cal-{id}" which sanitizes to "publish_cal_{id}".
type updatePublishAtSignals struct {
	PublishCal struct {
		DateValue string `json:"dateValue"`
	} `json:"publish_cal"`
}

func (h *AdminHandler) episodeUpdatePublishAt(c echo.Context) error {
	id := parseInt64(c.Param("id"))

	// Read the namespaced signal: the calendar ID is "publish-cal-{id}"
	// utils.Signals replaces "-" with "_" for JS compatibility
	signalKey := fmt.Sprintf("publish_cal_%d", id)

	var raw map[string]any
	if err := readSignals(c, &raw); err != nil {
		sse(c).MarshalAndPatchSignals(map[string]string{"error": "Invalid input"})
		return nil
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
			sse(c).MarshalAndPatchSignals(map[string]string{"error": "Invalid date format"})
			return nil
		}
		publishAt = &t
	}

	updated, err := h.episodeService.Update(c.Request().Context(), &domain.UpdateEpisode{
		ID:        id,
		PublishAt: publishAt,
	})
	if err != nil {
		log.Printf("update publish_at failed: %v", err)
		sse(c).MarshalAndPatchSignals(map[string]string{"error": "Failed to update publish date"})
		return nil
	}

	out := sse(c)
	rowID := fmt.Sprintf("episode-row-%d", id)
	out.PatchElementTempl(episodeview.EpisodeRow(updated), datastar.WithSelectorID(rowID), datastar.WithModeOuter())
	return nil
}

func (h *AdminHandler) episodeDelete(c echo.Context) error {
	id := parseInt64(c.Param("id"))

	if err := h.episodeService.Delete(c.Request().Context(), id); err != nil {
		log.Printf("delete episode failed: %v", err)
		sse(c).MarshalAndPatchSignals(map[string]string{"error": "Failed to delete episode"})
		return nil
	}

	out := sse(c)
	rowID := fmt.Sprintf("episode-row-%d", id)
	out.RemoveElementByID(rowID)
	return nil
}
