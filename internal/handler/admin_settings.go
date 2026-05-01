package handler

import (
	"crypto/rand"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/a-h/templ"
	"github.com/gracchi-stdio/castogo/internal/config"
	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/gracchi-stdio/castogo/internal/view"
	"github.com/labstack/echo/v4"
)

func (h *AdminHandler) settingsPage(c echo.Context) error {
	config, _ := h.settingsService.GetPodcastConfig(c.Request().Context())
	// config might be nil (first run) — template handles nil gracefully
	return echo.WrapHandler(templ.Handler(view.SettingsPage(getSharedData(c), config)))(c)
}

func (h *AdminHandler) settingsSave(c echo.Context) error {
	// Parse form data
	return nil
}

func (h *AdminHandler) settingsUploadCoverImage(c echo.Context) error {
	// Parse multipart form data (max 10MB)
	if err := c.Request().ParseMultipartForm(10 << 20); err != nil {
		sse(c).MarshalAndPatchSignals(map[string]string{
			"error":          "Failed to parse form data",
			"loading_status": "",
			"loading_msg":    "",
		})
		return nil
	}

	// read file
	file, header, err := c.Request().FormFile("cover_image")
	if err != nil {
		sse(c).MarshalAndPatchSignals(map[string]string{
			"error":          "Please select a file to upload",
			"loading_status": "",
			"loading_msg":    "",
		})
		return nil
	}
	defer file.Close()

	// Validate file type (basic check)
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		sse(c).MarshalAndPatchSignals(map[string]string{
			"error":          "Invalid file type. Please upload a JPG, JPEG, PNG, or WebP image.",
			"loading_status": "",
			"loading_msg":    "",
		})
		return nil
	}

	// generating unique filename
	b := make([]byte, 4)
	rand.Read(b)
	filename := fmt.Sprintf("%s/setting_cover_%x%s", strings.ToLower(strings.TrimSpace(config.Cfg.AppName)), b, ext)

	// upload
	sse(c).MarshalAndPatchSignals(map[string]string{"loading_msg": "Uploading cover image file..."})
	url, err := h.storageService.UploadFile(c.Request().Context(), file, filename)
	if err != nil {
		sse(c).MarshalAndPatchSignals(map[string]string{
			"error":          "Failed to upload image. Please try again.",
			"loading_status": "",
			"loading_msg":    "",
		})
		log.Printf("Error uploading file: %v", err)
		return nil
	}

	// save on database — only update the cover_image_url column
	sse(c).MarshalAndPatchSignals(map[string]string{"loading_msg": "Saving changes..."})
	settings := domain.UpdatePodcastConfig{
		ID:            parseInt64(c.FormValue("id")),
		CoverImageURL: &url,
	}
	if _, err := h.settingsService.UpdatePodcastConfig(c.Request().Context(), &settings); err != nil {
		sse(c).MarshalAndPatchSignals(map[string]string{
			"error":          "Failed to save cover image URL. Please try again.",
			"loading_status": "",
			"loading_msg":    "",
		})
		return nil
	}

	// Update frontend — set the CDN URL and clear uploading state
	sse(c).MarshalAndPatchSignals(map[string]string{
		"cover_image_url": url,
		"loading_status":  "",
		"loading_msg":     "",
	})
	return nil
}
