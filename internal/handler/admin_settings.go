package handler

import (
	"github.com/a-h/templ"
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

	return nil
}
