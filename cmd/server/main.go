package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
	"time"

	"github.com/a-h/templ"
	"github.com/go-playground/validator/v10"
	echosession "github.com/labstack/echo-contrib/v5/session"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	"github.com/gracchi-stdio/castogo/internal/config"
	"github.com/gracchi-stdio/castogo/internal/handler"
	"github.com/gracchi-stdio/castogo/internal/repository/postgres"
	"github.com/gracchi-stdio/castogo/internal/service"
	"github.com/gracchi-stdio/castogo/internal/session"
	"github.com/gracchi-stdio/castogo/internal/view/notfoundview"

	_ "github.com/gracchi-stdio/castogo/internal/view/editors/blockeditor/types" // initialize block types
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i any) error {
	if err := cv.validator.Struct(i); err != nil {
		return err
	}
	return nil
}

func main() {
	err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	validate := validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("form"), ",", 2)[0]
		if name == "-" || name == "" {
			return fld.Name
		}
		return name
	})

	e := echo.New()
	e.Validator = &CustomValidator{validator: validate}

	// Echo's router treats "/admin" and "/admin/" as distinct routes, so
	// trailing-slash URLs miss every registered route and fall through to 404.
	// Strip the trailing slash before routing so canonical (slash-less) URLs
	// resolve. Pre-middleware runs before the router; with no RedirectCode it
	// rewrites the path in place (no extra round-trip, safe for POST/Datastar)
	// and skips the root "/".
	e.Pre(middleware.RemoveTrailingSlash())

	// Database
	db, err := postgres.NewPool(context.Background(), config.Cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	skipHealth := func(c *echo.Context) bool {
		return c.Request().URL.Path == "/healthcheck"
	}

	// Session store (PostgreSQL-backed)
	sessionStore, err := session.NewPGStore(db, []byte(config.Cfg.SessionSecret))
	if err != nil {
		log.Fatalf("failed to create session store: %v", err)
	}

	// Middleware
	e.Use(middleware.Recover())
	e.Use(handler.RequestLogger(skipHealth))
	e.Use(echosession.Middleware(sessionStore))

	// Repositories
	userRepo := postgres.NewUserRepo(db)
	episodeRepo := postgres.NewEpisodeRepo(db)
	podcastRepo := postgres.NewPodcastConfigRepository(db)
	pageRepo := postgres.NewPageRepo(db)

	// Services
	authService := service.NewAuthService(userRepo)
	storageService := service.NewBunnyStorageService(config.Cfg.BunnyStorageEndpoint, config.Cfg.BunnyStoragePassword)
	episodeService := service.NewEpisodeService(episodeRepo)
	audioProcessor := service.NewFFmpegProcessor()
	settingsService := service.NewSettingsService(podcastRepo)
	feedService := service.NewFeedService(podcastRepo, episodeRepo)
	pageService := service.NewPageService(pageRepo, episodeRepo)

	// Handlers
	publicHandler := handler.NewPublicHandler(authService, feedService, pageService, episodeService, settingsService)
	publicHandler.RegisterRoutes(e)

	// Admin routes (protected by auth middleware)
	adminHandler := handler.NewAdminHandler(
		storageService,
		episodeService,
		audioProcessor,
		settingsService,
		pageService,
	)
	adminGroup := e.Group("/admin", handler.AuthMiddleware(userRepo))
	adminHandler.RegisterRoutes(adminGroup)

	e.GET("/healthcheck", func(c *echo.Context) error {
		if err := db.Ping(c.Request().Context()); err != nil {
			return c.JSON(503, map[string]string{"status": "unhealthy", "error": err.Error()})
		}
		return c.JSON(200, map[string]string{"db": "ok"})
	})

	// DefaultHTTPErrorHandler is a factory in v5 (no longer a method on *Echo);
	// capture the default handler once so the fallback below can call it.
	defaultErrHandler := echo.DefaultHTTPErrorHandler(false)
	e.HTTPErrorHandler = func(c *echo.Context, err error) {
		// In v5 echo.ErrNotFound is a distinct type (*httpError) from *echo.HTTPError,
		// so check the status code via the HTTPStatusCoder interface that both
		// implement, rather than asserting *echo.HTTPError.
		code := http.StatusInternalServerError
		if coder, ok := err.(echo.HTTPStatusCoder); ok {
			code = coder.StatusCode()
		}
		if code == http.StatusNotFound {
			// render the custom 404 templ template
			if resp, _ := echo.UnwrapResponse(c.Response()); resp != nil {
				resp.Status = http.StatusNotFound
			}
			echo.WrapHandler(templ.Handler(notfoundview.NotFoundView()))(c)
			return
		}
		// fall through to Echo's default error handler
		defaultErrHandler(c, err)
	}

	// Static files — use middleware (not route-based e.Static) so it falls through
	// to route handlers when no static file exists. e.Static would register GET /*
	// which overwrites the /*pageSlug wildcard route used by the page resolver.
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:   "public",
		Browse: false,
	}))

	httpClient := &http.Client{Timeout: 60 * time.Second} // 60 Timeout
	fetcher := service.NewBunnyLogFetcher(
		"https://logging.bunnycdn.com", // see Bunny.net docs for log fetching
		config.Cfg.BunnyAPIKey,
		config.Cfg.BunnyPullZoneID,
		httpClient, // use default HTTP client
	)
	analyticsRepo := postgres.NewAnalyticsPostgres(db)
	analyticsService := service.NewAnalyticsService(fetcher, analyticsRepo, episodeRepo)
	worker := service.NewAnalyticsWorker(analyticsService, 5*time.Hour)

	// graceful shutdown on SIGINT/SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	worker.Start(ctx)

	// In v5, e.Start traps SIGINT/SIGTERM itself and performs graceful HTTP
	// shutdown (default 10s GracefulTimeout) before returning — e.Shutdown()
	// was removed, so the manual signal/shutdown dance is no longer needed.
	if err := e.Start(":" + config.Cfg.Port); err != nil {
		e.Logger.Error("failed to start server", "error", err)
	}

	worker.Stop()

	db.Close()

	log.Println("Server gracefully stopped")
}
