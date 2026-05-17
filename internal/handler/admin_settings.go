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
	"github.com/gracchi-stdio/castogo/internal/view/settings_page"
	"github.com/labstack/echo/v4"
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

func (h *AdminHandler) settingsPage(c echo.Context) error {
	config, _ := h.settingsService.GetPodcastConfig(c.Request().Context())
	// config might be nil (first run) — template handles nil gracefully
	return echo.WrapHandler(templ.Handler(settings_page.SettingsPage(getSharedData(c), config)))(c)
}

type SettingsForm struct {
	ID            int64  `form:"id" validate:"required"`
	Title         string `form:"title" validate:"required"`
	Description   string `form:"description"`
	SiteURL       string `form:"site_url"`
	Language      string `form:"language"`
	Copyright     string `form:"copyright"`
	AuthorName    string `form:"author_name"`
	AuthorEmail   string `form:"author_email"`
	CoverImageURL string `form:"cover_image_url"`
	Category      string `form:"category"`
	Subcategory   string `form:"subcategory"`
	OwnerName     string `form:"owner_name"`
	OwnerEmail    string `form:"owner_email"`
}

func (h *AdminHandler) settingsSave(c echo.Context) error {
	settingInput := &SettingsForm{
		ID:            parseInt64(c.FormValue("id")),
		Title:         c.FormValue("title"),
		Description:   c.FormValue("description"),
		SiteURL:       c.FormValue("site_url"),
		Language:      c.FormValue("language"),
		Copyright:     c.FormValue("copyright"),
		AuthorName:    c.FormValue("author_name"),
		AuthorEmail:   c.FormValue("author_email"),
		CoverImageURL: c.FormValue("cover_image_url"),
		Category:      c.FormValue("category"),
		Subcategory:   c.FormValue("subcategory"),
		OwnerName:     c.FormValue("owner_name"),
		OwnerEmail:    c.FormValue("owner_email"),
	}
	if err := c.Validate(settingInput); err != nil {
		sse(c).MarshalAndPatchSignals(map[string]string{
			"error":          "Invalid form data",
			"loading_status": "",
			"loading_msg":    "",
		})
		return nil
	}
	if settingInput.Category != "" && !domain.IsValidCategory(settingInput.Category, settingInput.Subcategory) {
		sse(c).MarshalAndPatchSignals(map[string]string{
			"error":          "Please choose a valid category and subcategory",
			"loading_status": "",
			"loading_msg":    "",
		})
		return nil
	}

	var subcategory *string
	if settingInput.Category == "" {
		subcategory = nil
	} else if domain.HasSubcategories(settingInput.Category) {
		subcategory = stringPtrIfNonEmpty(settingInput.Subcategory)
	} else {
		subcategory = stringPtr("")
	}

	update := &domain.UpdatePodcastConfig{
		ID:            settingInput.ID,
		Title:         stringPtrIfNonEmpty(settingInput.Title),
		Description:   stringPtrIfNonEmpty(settingInput.Description),
		SiteURL:       stringPtrIfNonEmpty(settingInput.SiteURL),
		Language:      stringPtrIfNonEmpty(settingInput.Language),
		Copyright:     stringPtrIfNonEmpty(settingInput.Copyright),
		AuthorName:    stringPtrIfNonEmpty(settingInput.AuthorName),
		AuthorEmail:   stringPtrIfNonEmpty(settingInput.AuthorEmail),
		CoverImageURL: stringPtrIfNonEmpty(settingInput.CoverImageURL),
		Category:      stringPtrIfNonEmpty(settingInput.Category),
		Subcategory:   subcategory,
		OwnerName:     stringPtrIfNonEmpty(settingInput.OwnerName),
		OwnerEmail:    stringPtrIfNonEmpty(settingInput.OwnerEmail),
	}
	if _, err := h.settingsService.UpdatePodcastConfig(c.Request().Context(), update); err != nil {
		sse(c).MarshalAndPatchSignals(map[string]string{
			"error":          "Failed to save settings. Please try again.",
			"loading_status": "",
			"loading_msg":    "",
		})
		return nil
	}

	sse(c).MarshalAndPatchSignals(map[string]string{
		"loading_status": "",
		"loading_msg":    "",
	})
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
