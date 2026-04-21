package handler

import (
	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v3"
	"github.com/gracchi-stdio/castogo/internal/config"
	"github.com/gracchi-stdio/castogo/internal/service"
	"github.com/gracchi-stdio/castogo/internal/view"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{
		auth: auth,
	}
}

func (h *AuthHandler) RegisterRoutes(app *fiber.App) {
	if config.Cfg.RegistrationEnabled {
		app.Get("/register", templ.Handler(view.RegisterPage()))
	}
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	return nil
}

func (h *AuthHandler) Logout(c fiber.Ctx) error {
	return nil
}
