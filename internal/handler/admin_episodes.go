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
	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/gracchi-stdio/castogo/internal/service"
	"github.com/gracchi-stdio/castogo/internal/view"
	"github.com/labstack/echo/v4"
)

func (h *AdminHandler) episodesList(c echo.Context) error {
	return echo.WrapHandler(templ.Handler(view.EpisodesListPage(getSharedData(c))))(c)
}

func (h *AdminHandler) episodeCreatePage(c echo.Context) error {
	return echo.WrapHandler(templ.Handler(view.EpisodeNewPage(getSharedData(c))))(c)
}

type EpisodeInput struct {
	Title       string `validate:"required"`
	Description string `validate:""`
}

func (h *AdminHandler) episodeCreateAction(c echo.Context) error {
	// Parse multipart form (100MB max)
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

	// Read uploaded file
	file, header, err := c.Request().FormFile("audio_file")
	if err != nil {
		sse(c).MarshalAndPatchSignals(map[string]string{"audio_file_error": "Audio file is required", "uploading": ""})
		return nil
	}
	defer file.Close()

	// Validate file type
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".mp3" && ext != ".wav" && ext != ".m4a" && ext != ".ogg" && ext != ".flac" {
		sse(c).MarshalAndPatchSignals(map[string]string{"audio_file_error": "Unsupported file type. Use mp3, wav, m4a, ogg, or flac.", "uploading": ""})
		return nil
	}

	// --- Streaming progress starts here ---
	out := sse(c)

	// Step 1: Save upload to temp file
	out.MarshalAndPatchSignals(map[string]string{"uploading_status": "Validating file..."})

	tmp, err := service.NewTempFile(file, ext)
	if err != nil {
		sse(c).MarshalAndPatchSignals(map[string]string{"error": "Failed to create temporary file", "uploading": ""})
		return nil
	}
	defer tmp.Cleanup()

	// Step 2: Process with FFmpeg
	out.MarshalAndPatchSignals(map[string]string{"uploading_status": "Processing audio..."})

	processResult, err := h.audioProcessor.Process(c.Request().Context(), tmp.Path, service.DefaultProcessingOptions())
	if err != nil {
		sse(c).MarshalAndPatchSignals(map[string]string{"error": "Failed to process audio file", "uploading": ""})
		return nil
	}
	defer os.Remove(processResult.OutputPath)

	// Step 3: Upload processed file to Bunny Storage
	out.MarshalAndPatchSignals(map[string]string{"uploading_status": "Uploading audio..."})

	slug := slugify(input.Title)
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

	// Step 4: Create episode record
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
		Status:         domain.EpisodeStatusDraft,
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

	// Done — navigate to episodes list
	out.MarshalAndPatchSignals(map[string]string{"uploading": ""})
	out.ExecuteScript("window.navigateAdmin('/admin/episodes')")
	return nil
}
