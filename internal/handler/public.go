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
	"github.com/gracchi-stdio/castogo/internal/view/searchview"
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

	// Search (before wildcard)
	e.GET("/search", h.searchPage)

	// RSS Feed (public)
	e.GET("/feed/podcast.xml", h.RSSFeed)

	// Auth
	e.GET("/login", echo.WrapHandler(templ.Handler(authview.LoginPage())))
	e.POST("/login", h.login)
	e.POST("/logout", h.logout)

	if config.Cfg.RegistrationEnabled {
		e.GET("/register", echo.WrapHandler(templ.Handler(authview.RegisterPage())))
	}

	e.GET("/*", h.pageResolver)
}

func (h *PublicHandler) homePage(c echo.Context) error {
	// Homepage is the root page — the one with empty slug and path ""
	page, err := h.pageService.GetPageByPath(c.Request().Context(), "")
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

	episodes, err := h.episodesService.ListPublishedWithPagePath(c.Request().Context(), 5, 0)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching latest episodes")
	}

	settings, err := h.settingsService.GetPodcastConfig(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching podcast settings")
	}

	data := &pageview.PageData{
		Page:     pwb.Page,
		Blocks:   toBlocks(pwb.Blocks),
		Episodes: episodes,
		Nav:      h.buildPublicNav(c),
		Settings: *settings,
	}

	return echo.WrapHandler(templ.Handler(pageview.PageView(data)))(c)
}

func (h *PublicHandler) pageResolver(c echo.Context) error {
	// Echo stores wildcard params as "*", not the named version
	slug := strings.Trim(c.Param("*"), "/")

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

	episodes, err := h.episodesService.ListPublishedWithPagePath(c.Request().Context(), 20, 0)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching latest episodes")
	}

	settings, err := h.settingsService.GetPodcastConfig(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching podcast settings")
	}

	data := &pageview.PageData{
		Page:     pwb.Page,
		Blocks:   toBlocks(pwb.Blocks),
		Episodes: episodes,
		Nav:      h.buildPublicNav(c),
		Settings: *settings,
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
		sse(c).MarshalAndPatchSignals(fieldValidationErrors(err, input))
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

func (h *PublicHandler) buildPublicNav(c echo.Context) *layout.PublicLayoutData {
	pages, err := h.pageService.GetTopLevelPublished(c.Request().Context())
	if err != nil {
		pages = nil
	}

	settings, err := h.settingsService.GetPodcastConfig(c.Request().Context())
	if err != nil {
		settings = nil
	}

	var navLinks []layout.NavLink
	for _, p := range pages {
		url := "/" + p.Path
		if p.Path == "" {
			url = "/"
		}
		navLinks = append(navLinks, layout.NavLink{
			Label: p.Title,
			URL:   url,
		})
	}

	navData := &layout.PublicNavbarData{
		NavLinks: navLinks,
	}
	data := &layout.PublicLayoutData{NavData: navData}
	if settings != nil {
		data.Settings = &layout.SettingsArgs{Title: settings.Title}
		navData.Title = settings.Title
	}
	return data
}

func (h *PublicHandler) searchPage(c echo.Context) error {
	query := strings.TrimSpace(c.QueryParam("q"))
	searchType := c.QueryParam("type")
	if searchType == "" {
		searchType = "all"
	}

	var pages []*domain.Page
	var episodes []*domain.Episode

	if query != "" {
		if searchType == "all" || searchType == "pages" {
			results, err := h.pageService.SearchPublished(c.Request().Context(), query, 20, 0)
			if err == nil {
				pages = results
			}
		}
		if searchType == "all" || searchType == "episodes" {
			results, err := h.episodesService.SearchPublished(c.Request().Context(), query, 20, 0)
			if err == nil {
				episodes = results
			}
		}
	}

	data := &searchview.SearchPageData{
		Query:    query,
		Type:     searchType,
		Pages:    pages,
		Episodes: episodes,
		Nav:      h.buildPublicNav(c),
	}

	return echo.WrapHandler(templ.Handler(searchview.SearchPage(data)))(c)
}

func toBlocks(blocks []*domain.PageBlock) []domain.PageBlock {
	result := make([]domain.PageBlock, len(blocks))
	for i, b := range blocks {
		result[i] = *b
	}
	return result
}
