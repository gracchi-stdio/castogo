package handler

import (
	"log"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/gracchi-stdio/castogo/internal/config"
	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/gracchi-stdio/castogo/internal/service"
	"github.com/gracchi-stdio/castogo/internal/view/authview"
	"github.com/gracchi-stdio/castogo/internal/view/layout"
	"github.com/gracchi-stdio/castogo/internal/view/pageview"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type PublicHandler struct {
	auth            *service.AuthService
	feedService     *service.FeedService
	pageService     *service.PageService
	episodesService *service.EpisodeService
	settingsService *service.SettingsService
}

func NewPublicHandler(
	auth *service.AuthService,
	feedService *service.FeedService,
	pageService *service.PageService,
	episodesService *service.EpisodeService,
	settingsService *service.SettingsService,
) *PublicHandler {
	return &PublicHandler{
		auth:            auth,
		feedService:     feedService,
		pageService:     pageService,
		episodesService: episodesService,
		settingsService: settingsService,
	}
}

func (h *PublicHandler) RegisterRoutes(e *echo.Echo) {
	// Public home page
	e.GET("/", h.homePage)

	// RSS Feed (public)
	e.GET("/feed/podcast.xml", h.RSSFeed)

	// Auth
	e.GET("/login", echo.WrapHandler(templ.Handler(authview.LoginPage())))
	e.POST("/login", h.login)
	e.POST("/logout", h.logout)

	if config.Cfg.RegistrationEnabled {
		e.GET("/register", echo.WrapHandler(templ.Handler(authview.RegisterPage())))
	}

	e.GET("/*pageSlug", h.pageResolver)
}

func (h *PublicHandler) homePage(c echo.Context) error {
	cfg, err := h.settingsService.GetPodcastConfig(c.Request().Context())
	if err != nil || cfg.HomepageID == nil {
		log.Printf("no homepage configured: %v", err)
		return echo.ErrNotFound
	}

	pwb, err := h.pageService.GetPageWithBlocks(c.Request().Context(), *cfg.HomepageID)
	if err != nil {
		log.Printf("error fetching home page: %v", err)
		return echo.ErrNotFound
	}

	episodes, err := h.episodesService.ListPublished(c.Request().Context(), 5, 0)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching latest episodes")
	}

	data := &pageview.PageData{
		Page:     pwb.Page,
		Blocks:   toBlocks(pwb.Blocks),
		Episodes: episodes,
		Nav:      buildPublicNav(c),
	}

	return echo.WrapHandler(templ.Handler(pageview.PageView(data)))(c)
}

func (h *PublicHandler) pageResolver(c echo.Context) error {
	slug := c.Param("pageSlug")
	slug = strings.Trim(slug, "/")

	page, err := h.pageService.GetPageByPath(c.Request().Context(), slug)
	if err != nil {
		return echo.ErrNotFound
	}

	if !page.IsPublished {
		return echo.ErrNotFound
	}

	pwb, err := h.pageService.GetPageWithBlocks(c.Request().Context(), page.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching page content")
	}

	episodes, err := h.episodesService.ListPublished(c.Request().Context(), 20, 0)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching latest episodes")
	}

	data := &pageview.PageData{
		Page:     pwb.Page,
		Blocks:   toBlocks(pwb.Blocks),
		Episodes: episodes,
		Nav:      buildPublicNav(c),
	}

	return echo.WrapHandler(templ.Handler(pageview.PageView(data)))(c)
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

func buildPublicNav(c echo.Context) *layout.PublicLayoutData {
	return &layout.PublicLayoutData{
		NavData: &layout.PublicNavbarData{
			NavLinks: []layout.NavLink{
				{Label: "Home", URL: "/"},
				{Label: "Episodes", URL: "/episodes"},
			},
		},
	}
}

func toBlocks(blocks []*domain.PageBlock) []domain.PageBlock {
	result := make([]domain.PageBlock, len(blocks))
	for i, b := range blocks {
		result[i] = *b
	}
	return result
}
