package handler

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/gracchi-stdio/castogo/internal/config"
	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/gracchi-stdio/castogo/internal/view/settingview"
	"github.com/labstack/echo/v5"
)

func stringPtrIfNonEmpty(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func stringPtr(value string) *string {
	return &value
}

func (h *AdminHandler) settingsPage(c *echo.Context) error {
	cfg, err := h.settingsService.GetPodcastConfig(c.Request().Context())
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load settings")
	}
	// cfg is nil on first run (no config row yet) — the template handles nil.
	return echo.WrapHandler(templ.Handler(settingview.SettingsPage(getSharedData(c), cfg)))(c)
}

func (h *AdminHandler) settingsSave(c *echo.Context) error {
	var raw settingsInput
	if err := readSignals(c, &raw); err != nil {
		return toast(c, "Invalid request", "error")
	}
	if err := validate.Struct(raw); err != nil {
		return patchFieldErrors(c, err, raw)
	}
	if raw.Category != "" && !domain.IsValidCategory(raw.Category, raw.Subcategory) {
		return toast(c, "Please choose a valid category and subcategory", "error")
	}

	var subcategory *string
	if raw.Category == "" {
		subcategory = nil
	} else if domain.HasSubcategories(raw.Category) {
		subcategory = stringPtrIfNonEmpty(raw.Subcategory)
	} else {
		subcategory = stringPtr("")
	}

	update := &domain.UpdatePodcastConfig{
		ID:            raw.ID,
		Title:         stringPtrIfNonEmpty(raw.Title),
		Description:   stringPtrIfNonEmpty(raw.Description),
		SiteURL:       stringPtrIfNonEmpty(raw.SiteURL),
		Language:      stringPtrIfNonEmpty(raw.Language),
		Copyright:     stringPtrIfNonEmpty(raw.Copyright),
		AuthorName:    stringPtrIfNonEmpty(raw.AuthorName),
		AuthorEmail:   stringPtrIfNonEmpty(raw.AuthorEmail),
		CoverImageURL: stringPtrIfNonEmpty(raw.CoverImageURL),
		Category:      stringPtrIfNonEmpty(raw.Category),
		Subcategory:   subcategory,
		OwnerName:     stringPtrIfNonEmpty(raw.OwnerName),
		OwnerEmail:    stringPtrIfNonEmpty(raw.OwnerEmail),
	}
	if _, err := h.settingsService.UpdatePodcastConfig(c.Request().Context(), update); err != nil {
		return toast(c, "Failed to save settings. Please try again.", "error")
	}
	return toast(c, "Settings saved successfully", "success")
}

func (h *AdminHandler) settingsUploadCoverImage(c *echo.Context) error {
	// Parse multipart form data (max 10MB)
	if err := c.Request().ParseMultipartForm(10 << 20); err != nil {
		out := sse(c)
		out.MarshalAndPatchSignals(map[string]string{
			"cover_uploading": "false",
			"cover_status":    "",
		})
		out.ExecuteScript(toastScript("Failed to parse form data", "error"))
		return nil
	}

	// Validate the config id up front so we don't upload a cover we can't attach.
	configID, err := strconv.ParseInt(c.FormValue("id"), 10, 64)
	if err != nil {
		out := sse(c)
		out.MarshalAndPatchSignals(map[string]string{
			"cover_uploading": "false",
			"cover_status":    "",
		})
		out.ExecuteScript(toastScript("Invalid settings ID", "error"))
		return nil
	}

	// read file
	file, header, err := c.Request().FormFile("cover_image")
	if err != nil {
		out := sse(c)
		out.MarshalAndPatchSignals(map[string]string{
			"cover_uploading": "false",
			"cover_status":    "",
		})
		out.ExecuteScript(toastScript("Please select a file to upload", "error"))
		return nil
	}
	defer file.Close()

	// Validate file type (basic check)
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		out := sse(c)
		out.MarshalAndPatchSignals(map[string]string{
			"cover_uploading": "false",
			"cover_status":    "",
		})
		out.ExecuteScript(toastScript("Invalid file type. Please upload a JPG, JPEG, PNG, or WebP image.", "error"))
		return nil
	}

	// generating unique filename
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	filename := fmt.Sprintf("%s/setting_cover_%x%s", strings.ToLower(strings.TrimSpace(config.Cfg.AppName)), b, ext)

	// upload
	sse(c).MarshalAndPatchSignals(map[string]string{"cover_status": "Uploading cover image file..."})
	url, err := h.storageService.UploadFile(c.Request().Context(), file, filename)
	if err != nil {
		out := sse(c)
		out.MarshalAndPatchSignals(map[string]string{
			"cover_uploading": "false",
			"cover_status":    "",
		})
		out.ExecuteScript(toastScript("Failed to upload image. Please try again.", "error"))
		log.Printf("Error uploading file: %v", err)
		return nil
	}

	// save on database — only update the cover_image_url column
	sse(c).MarshalAndPatchSignals(map[string]string{"cover_status": "Saving changes..."})
	settings := domain.UpdatePodcastConfig{
		ID:            configID,
		CoverImageURL: &url,
	}
	if _, err := h.settingsService.UpdatePodcastConfig(c.Request().Context(), &settings); err != nil {
		out := sse(c)
		out.MarshalAndPatchSignals(map[string]string{
			"cover_uploading": "false",
			"cover_status":    "",
		})
		out.ExecuteScript(toastScript("Failed to save cover image URL. Please try again.", "error"))
		return nil
	}

	// Update frontend — set the CDN URL and clear uploading state
	sse(c).MarshalAndPatchSignals(map[string]string{
		"cover_image_url":  url,
		"cover_uploading":  "false",
		"cover_status":     "",
	})
	return nil
}
