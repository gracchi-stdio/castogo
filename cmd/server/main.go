package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/go-playground/validator/v10"
	echosession "github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

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
	e.HideBanner = false

	// Database
	db, err := postgres.NewPool(context.Background(), config.Cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	skipHealth := func(c echo.Context) bool {
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

	e.GET("/healthcheck", func(c echo.Context) error {
		if err := db.Ping(c.Request().Context()); err != nil {
			return c.JSON(503, map[string]string{"status": "unhealthy", "error": err.Error()})
		}
		return c.JSON(200, map[string]string{"db": "ok"})
	})

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if he, ok := err.(*echo.HTTPError); ok && he.Code == 404 {
			// render your custom 404 templ template
			c.Response().Status = http.StatusNotFound
			echo.WrapHandler(templ.Handler(notfoundview.NotFoundView()))(c)
			return
		}
		// fall through to Echo's default error handler
		e.DefaultHTTPErrorHandler(err, c)
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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	worker.Start(ctx)

	go func() {
		if err := e.Start(":" + config.Cfg.Port); err != nil {
			e.Logger.Error("failed to start server", "error", err)
		}
	}()

	// Block until we receive a shutdown signal (e.g. Ctrl+C)
	<-ctx.Done()
	log.Println("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	e.Shutdown(shutdownCtx)

	worker.Stop()

	db.Close()

	log.Println("Server gracefully stopped")
}
