package handler

import (
	"encoding/json"
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gracchi-stdio/castogo/internal/config"
	"github.com/gracchi-stdio/castogo/internal/service"
	"github.com/gracchi-stdio/castogo/internal/view"
	"github.com/starfederation/datastar-go/datastar"
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
	app.Post("/login", adaptor.HTTPHandlerFunc(h.loginHTTP))

	if config.Cfg.RegistrationEnabled {
		app.Get("/register", templ.Handler(view.RegisterPage()))
	}
}

type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

func (h *AuthHandler) loginHTTP(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)

	input := new(LoginInput)
	if err := json.NewDecoder(r.Body).Decode(input); err != nil {
		sse.MarshalAndPatchSignals(map[string]string{"error": "Invalid request"})
		return
	}

	validate := validator.New()
	if err := validate.Struct(input); err != nil {
		sse.MarshalAndPatchSignals(fieldValidationErrors(err))
		return
	}

	user, err := h.auth.VerifyCredentials(r.Context(), input.Email, input.Password)
	if err != nil {
		sse.MarshalAndPatchSignals(map[string]string{"error": "Invalid email or password"})
		return
	}

	// TODO: set session with user.ID
	_ = user

	sse.ExecuteScript("window.location.href = '/admin'")
}

func (h *AuthHandler) Logout(c fiber.Ctx) error {
	return nil
}
