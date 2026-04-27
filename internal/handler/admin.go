package handler

import (
	"crypto/rand"
	"fmt"
	"log"
	"time"
	"path/filepath"
	"strings"

	"github.com/a-h/templ"
	"github.com/gracchi-stdio/castogo/internal/service"
	"github.com/gracchi-stdio/castogo/internal/view"
	"github.com/labstack/echo/v4"
)

type AdminHandler struct {
	storageService service.StorageService
}

func NewAdminHandler(storageService service.StorageService) *AdminHandler {
	return &AdminHandler{
		storageService: storageService,
	}
}

func (h *AdminHandler) RegisterRoutes(g *echo.Group) {
	g.GET("", h.dashboard)
	g.GET("/episodes", h.episodesList)
	g.GET("/episodes/create", h.episodeCreatePage)
	g.POST("/episodes/create", h.episodeCreateAction)
}

func (h *AdminHandler) dashboard(c echo.Context) error {
	return echo.WrapHandler(templ.Handler(view.DashboardPage(getSharedData(c))))(c)
}

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

	// Step 1: Upload audio file
	out.MarshalAndPatchSignals(map[string]string{"uploading_status": "Validating file..."})
	time.Sleep(1 * time.Second)

	out.MarshalAndPatchSignals(map[string]string{"uploading_status": "Uploading audio file..."})

	slug := slugify(input.Title)
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	uniqueID := fmt.Sprintf("%x", b)
	filename := fmt.Sprintf("episodes/%s-%s%s", slug, uniqueID[:8], ext)

	cdnURL, err := h.storageService.UploadFile(c.Request().Context(), file, filename)
	if err != nil {
		log.Printf("upload failed: %v", err)
		out.MarshalAndPatchSignals(map[string]string{"error": "Failed to upload audio file", "uploading": ""})
		return nil
	}

	// Step 2: Create episode record
	out.MarshalAndPatchSignals(map[string]string{"uploading_status": "Creating episode record..."})
	time.Sleep(1 * time.Second)
	// TODO: save episode to DB via episodeService

	out.MarshalAndPatchSignals(map[string]string{"uploading_status": "Finishing up..."})
	time.Sleep(500 * time.Millisecond)

	log.Printf("Episode created: title=%s audio=%s", input.Title, cdnURL)

	// Done — navigate to episodes list
	out.MarshalAndPatchSignals(map[string]string{"uploading": ""})
	out.ExecuteScript("window.navigateAdmin('/admin/episodes')")
	return nil
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			return r
		}
		if r == ' ' || r == '-' || r == '_' {
			return '-'
		}
		return -1
	}, s)
	return strings.Trim(s, "-")
}
