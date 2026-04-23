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
}

func (h *AdminHandler) dashboard(c echo.Context) error {
	return echo.WrapHandler(templ.Handler(view.DashboardPage(getSharedData(c))))(c)
}
