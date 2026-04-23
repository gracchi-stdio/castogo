package handler

import (
	"log"

	"github.com/a-h/templ"
	"github.com/gracchi-stdio/castogo/internal/config"
	"github.com/gracchi-stdio/castogo/internal/service"
	"github.com/gracchi-stdio/castogo/internal/view"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) RegisterRoutes(e *echo.Echo) {
	e.GET("/login", echo.WrapHandler(templ.Handler(view.LoginPage())))
	e.POST("/login", h.login)
	e.POST("/logout", h.logout)

	if config.Cfg.RegistrationEnabled {
		e.GET("/register", echo.WrapHandler(templ.Handler(view.RegisterPage())))
	}
}

type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

func (h *AuthHandler) login(c echo.Context) error {
	input := new(LoginInput)
	if err := readSignals(c, input); err != nil {
		sse(c).MarshalAndPatchSignals(map[string]string{"error": "Invalid request"})
		return nil
	}

	if err := validate.Struct(input); err != nil {
		sse(c).MarshalAndPatchSignals(fieldValidationErrors(err))
		return nil
	}

	user, err := h.auth.VerifyCredentials(c.Request().Context(), input.Email, input.Password)
	if err != nil {
		sse(c).MarshalAndPatchSignals(map[string]string{"error": "Invalid email or password"})
		return nil
	}

	sess, _ := session.Get("session", c)
	sess.Values["user_id"] = user.ID.String()

	if err := sess.Save(c.Request(), c.Response().Writer); err != nil {
		log.Printf("session save error: %v", err)
	}

	sse(c).Redirect("/admin")
	return nil
}

func (h *AuthHandler) logout(c echo.Context) error {
	sess, _ := session.Get("session", c)
	sess.Options.MaxAge = -1
	sess.Save(c.Request(), c.Response().Writer)
	sse(c).Redirect("/login")
	return nil
}
