package handler

import (
	"github.com/a-h/templ"
	"github.com/gracchi-stdio/castogo/internal/view"
	"github.com/labstack/echo/v4"
)

type AdminHandler struct{}

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{}
}

func (h *AdminHandler) RegisterRoutes(g *echo.Group) {
	g.GET("", h.dashboard)
	g.GET("/episodes", h.episodesList)
	g.GET("/episodes/new", h.episodeNew)
}

func (h *AdminHandler) dashboard(c echo.Context) error {
	return echo.WrapHandler(templ.Handler(view.DashboardPage(getSharedData(c))))(c)
}

func (h *AdminHandler) episodesList(c echo.Context) error {
	return echo.WrapHandler(templ.Handler(view.EpisodesListPage(getSharedData(c))))(c)
}

func (h *AdminHandler) episodeNew(c echo.Context) error {
	return echo.WrapHandler(templ.Handler(view.EpisodeNewPage(getSharedData(c))))(c)
}
