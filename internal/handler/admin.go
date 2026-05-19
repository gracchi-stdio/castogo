package handler

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/gracchi-stdio/castogo/internal/service"
	"github.com/gracchi-stdio/castogo/internal/view/dashboardview"
	"github.com/labstack/echo/v4"
)

type AdminHandler struct {
	storageService  service.StorageService
	episodeService  *service.EpisodeService
	audioProcessor  service.AudioProcessor
	settingsService *service.SettingsService
	landingService  *service.LandingPageService
}

func NewAdminHandler(
	storageService service.StorageService,
	episodeService *service.EpisodeService,
	audioProcessor service.AudioProcessor,
	settingsService *service.SettingsService,
	landingService *service.LandingPageService) *AdminHandler {
	return &AdminHandler{
		storageService:  storageService,
		episodeService:  episodeService,
		audioProcessor:  audioProcessor,
		settingsService: settingsService,
		landingService:  landingService,
	}
}

func (h *AdminHandler) RegisterRoutes(g *echo.Group) {
	g.GET("", h.dashboard)

	// episodes
	g.GET("/episodes", h.episodesList)
	g.GET("/episodes/create", h.episodeCreatePage)
	g.POST("/episodes/create", h.episodeCreateAction)
	g.PATCH("/episodes/:id/publish-at", h.episodeUpdatePublishAt)
	g.DELETE("/episodes/:id", h.episodeDelete)

	// settings
	g.GET("/settings", h.settingsPage)
	g.POST("/settings", h.settingsSave)
	g.POST("/settings/upload-cover", h.settingsUploadCoverImage)
	g.GET("/settings/subcategories", h.subcategoriesSSE)

	// landing page
	g.GET("/landing", h.landingEditor)
	g.POST("/landing/:sectionKey", h.landingSaveSection)
}

func (h *AdminHandler) dashboard(c echo.Context) error {
	stats, err := h.episodeService.GetDashboardStats(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load dashboard stats")
	}

	args := dashboardview.DashboardPageArgs{
		Total:     stats.Total,
		Published: stats.Published,
		Drafts:    stats.Drafts,
		Scheduled: stats.Scheduled,
	}

	return echo.WrapHandler(templ.Handler(dashboardview.DashboardPage(getSharedData(c), args)))(c)
}
