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
	pageService     *service.PageService
}

func NewAdminHandler(
	storageService service.StorageService,
	episodeService *service.EpisodeService,
	audioProcessor service.AudioProcessor,
	settingsService *service.SettingsService,
	pageService *service.PageService,
) *AdminHandler {
	return &AdminHandler{
		storageService:  storageService,
		episodeService:  episodeService,
		audioProcessor:  audioProcessor,
		settingsService: settingsService,
		pageService:     pageService,
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

	// pages
	g.GET("/pages", h.pageList)
	g.GET("/pages/create", h.pageCreate)
	g.POST("/pages/create", h.pageCreateAction)
	g.GET("/pages/:id/edit", h.pageEdit)
	g.PATCH("/pages/:id", h.pageUpdateAction)
	g.DELETE("/pages/:id", h.pageDeleteAction)
	g.POST("/pages/blocks/upload-image", h.blockUploadImage)
	g.POST("/pages/:id/blocks", h.blockCreateAction)
	g.PATCH("/pages/:id/blocks/reorder", h.blockReorderAction)
	g.PATCH("/pages/:id/blocks/:blockId", h.blockUpdateAction)
	g.DELETE("/pages/:id/blocks/:blockId", h.blockDeleteAction)
	g.POST("/pages/:id/blocks/:blockId/items", h.blockAddItemAction)
	g.DELETE("/pages/:id/blocks/:blockId/items/:index", h.blockRemoveItemAction)
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
