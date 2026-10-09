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
		if !domain.CategoryExists(raw.Category) {
			return patchSignals(c, map[string]string{"category_error": "Choose a valid category"})
		}
		return patchSignals(c, map[string]string{"subcategory_error": "Choose a valid subcategory for this category"})
	}

	// The form posts the whole signal store, so it is the source of truth:
	// every field is written as-is — an empty string clears the stored value
	// (nil would mean "keep old" under the COALESCE partial update).
	// Subcategory only carries meaning under categories that have subcategories.
	subcategory := new("")
	if domain.HasSubcategories(raw.Category) {
		subcategory = new(raw.Subcategory)
	}

	update := &domain.UpdatePodcastConfig{
		ID:            raw.ID,
		Title:         new(raw.Title),
		Description:   new(raw.Description),
		SiteURL:       new(raw.SiteURL),
		Language:      new(raw.Language),
		Copyright:     new(raw.Copyright),
		AuthorName:    new(raw.AuthorName),
		AuthorEmail:   new(raw.AuthorEmail),
		CoverImageURL: new(raw.CoverImageURL),
		Category:      new(raw.Category),
		Subcategory:   subcategory,
		OwnerName:     new(raw.OwnerName),
		OwnerEmail:    new(raw.OwnerEmail),
	}
	if _, err := h.settingsService.UpdatePodcastConfig(c.Request().Context(), update); err != nil {
		return toast(c, "Failed to save settings. Please try again.", "error")
	}
	bustCache(c, "/admin/settings")
	return toast(c, "Settings saved successfully", "success")
}

func (h *AdminHandler) settingsUploadCoverImage(c *echo.Context) error {
	// coverUploadFail resets the upload UI state and surfaces a safe message.
	coverUploadFail := func(category, msg string, cause error) error {
		if cause != nil {
			log.Printf("settings cover upload: %s: %v", category, cause)
		} else {
			log.Printf("settings cover upload: %s", category)
		}
		patchSignals(c, map[string]string{
			"cover_uploading": "",
			"cover_status":    "",
		})
		return toast(c, msg, "error")
	}

	// Parse multipart form data (max 10MB)
	if err := c.Request().ParseMultipartForm(10 << 20); err != nil {
		return coverUploadFail("parse multipart form", "The cover image upload could not be read. Please try again.", err)
	}

	// Validate the config id up front so we don't upload a cover we can't attach.
	configID, err := strconv.ParseInt(c.FormValue("id"), 10, 64)
	if err != nil {
		return coverUploadFail("parse settings ID", "The settings record could not be identified.", err)
	}

	file, header, err := c.Request().FormFile("image_file")
	if err != nil {
		return coverUploadFail("read image file", "Please select a cover image to upload.", err)
	}
	defer file.Close()

	// Validate file type (basic check)
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		return coverUploadFail("validate image type", "Please upload a JPG, JPEG, PNG, or WebP image.", nil)
	}

	// generating unique filename
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	filename := fmt.Sprintf("%s/setting_cover_%x%s", strings.ToLower(strings.TrimSpace(config.Cfg.AppName)), b, ext)

	// upload
	patchSignals(c, map[string]string{"cover_status": "Uploading cover image file..."})
	url, err := h.storageService.UploadFile(c.Request().Context(), file, filename)
	if err != nil {
		return coverUploadFail("store image", "The cover image could not be uploaded. Please try again.", err)
	}

	// save on database — only update the cover_image_url column
	patchSignals(c, map[string]string{"cover_status": "Saving changes..."})
	settings := domain.UpdatePodcastConfig{
		ID:            configID,
		CoverImageURL: &url,
	}
	if _, err := h.settingsService.UpdatePodcastConfig(c.Request().Context(), &settings); err != nil {
		return coverUploadFail("save image URL", "The cover image uploaded but could not be saved. Please try again.", err)
	}

	// Update frontend — set the CDN URL and clear uploading state
	bustCache(c, "/admin/settings")
	patchSignals(c, map[string]string{
		"cover_image_url": url,
		"cover_uploading": "",
		"cover_status":    "",
	})
	return nil
}
