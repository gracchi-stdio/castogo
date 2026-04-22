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
	app.Get("/login", templ.Handler(view.LoginPage()))
	app.Post("/login", h.Login)

	if config.Cfg.RegistrationEnabled {
		app.Get("/register", templ.Handler(view.RegisterPage()))
	}
}

type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	input := new(LoginInput)
	if err := c.Bind().Body(input); err != nil {
		writeSignals(c, fieldValidationErrors(err))
		return nil
	}

	user, err := h.auth.VerifyCredentials(c.Context(), input.Email, input.Password)
	if err != nil {
		writeError(c, "Invalid email or password")
		return nil
	}

	// TODO: set session with user.ID
	_ = user

	writeSSE(c, "datastar-execute-script", "script window.location.href = '/'")
	return nil
}

func (h *AuthHandler) Logout(c fiber.Ctx) error {
	return nil
}
