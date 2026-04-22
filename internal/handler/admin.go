package handler

import (
	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v3"
	"github.com/gracchi-stdio/castogo/internal/view"
)

type AdminHandler struct{}

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{}
}

func (h *AdminHandler) RegisterRoutes(app *fiber.App) {
	app.Get("/admin", templ.Handler(view.DashboardPage()))
}
