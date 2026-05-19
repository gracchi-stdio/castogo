package handler

import (
	"log"

	"github.com/a-h/templ"
	"github.com/gracchi-stdio/castogo/internal/config"
	"github.com/gracchi-stdio/castogo/internal/service"
	"github.com/gracchi-stdio/castogo/internal/view/authview"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type PublicHandler struct {
	auth        *service.AuthService
	feedService *service.FeedService
}

func NewPublicHandler(auth *service.AuthService, feedService *service.FeedService) *PublicHandler {
	return &PublicHandler{auth: auth, feedService: feedService}
}

func (h *PublicHandler) RegisterRoutes(e *echo.Echo) {
	// RSS Feed (public)
	e.GET("/feed/podcast.xml", h.RSSFeed)

	// Auth
	e.GET("/login", echo.WrapHandler(templ.Handler(authview.LoginPage())))
	e.POST("/login", h.login)
	e.POST("/logout", h.logout)

	if config.Cfg.RegistrationEnabled {
		e.GET("/register", echo.WrapHandler(templ.Handler(authview.RegisterPage())))
	}
}

type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

func (h *PublicHandler) login(c echo.Context) error {
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

func (h *PublicHandler) logout(c echo.Context) error {
	sess, _ := session.Get("session", c)
	sess.Options.MaxAge = -1
	sess.Save(c.Request(), c.Response().Writer)
	sse(c).Redirect("/login")
	return nil
}
