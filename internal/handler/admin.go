package handler

import (
	"github.com/a-h/templ"
	"github.com/gracchi-stdio/castogo/internal/service"
	"github.com/gracchi-stdio/castogo/internal/view"
	"github.com/labstack/echo/v4"
)

type AdminHandler struct {
	storageService  service.StorageService
	episodeService  *service.EpisodeService
	audioProcessor  service.AudioProcessor
	settingsService *service.SettingsService
}

func NewAdminHandler(
	storageService service.StorageService,
	episodeService *service.EpisodeService,
	audioProcessor service.AudioProcessor,
	settingsService *service.SettingsService) *AdminHandler {
	return &AdminHandler{
		storageService:  storageService,
		episodeService:  episodeService,
		audioProcessor:  audioProcessor,
		settingsService: settingsService,
	}
}

func (h *AdminHandler) RegisterRoutes(g *echo.Group) {
	g.GET("", h.dashboard)
	g.GET("/episodes", h.episodesList)
	g.GET("/episodes/create", h.episodeCreatePage)
	g.POST("/episodes/create", h.episodeCreateAction)

	// settings
	g.GET("/settings", h.settingsPage)
	g.POST("/settings", h.settingsSave)
	g.POST("/settings/upload-cover", h.settingsUploadCoverImage)
		g.GET("/settings/subcategories", h.subcategoriesSSE)

}

func (h *AdminHandler) dashboard(c echo.Context) error {
	return echo.WrapHandler(templ.Handler(view.DashboardPage(getSharedData(c))))(c)
}
