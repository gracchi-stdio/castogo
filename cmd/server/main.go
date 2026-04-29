package main

import (
	"context"
	"log"

	echosession "github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/gracchi-stdio/castogo/internal/config"
	"github.com/gracchi-stdio/castogo/internal/handler"
	"github.com/gracchi-stdio/castogo/internal/repository/postgres"
	"github.com/gracchi-stdio/castogo/internal/service"
	"github.com/gracchi-stdio/castogo/internal/session"
)

func main() {
	err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	e := echo.New()
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

	// Services
	userRepo := postgres.NewUserRepo(db)
	authService := service.NewAuthService(userRepo)
	storageService := service.NewBunnyStorageService(config.Cfg.BunnyStorageEndpoint, config.Cfg.BunnyStoragePassword)
	episodeRepo := postgres.NewEpisodeRepo(db)
	episodeService := service.NewEpisodeService(episodeRepo)
	audioProcessor := service.NewFFmpegProcessor()
	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	authHandler.RegisterRoutes(e)

	clockHandler := handler.NewClockHandler()
	clockHandler.RegisterRoutes(e)

	filterHandler := handler.NewFilterHandler()
	filterHandler.RegisterRoutes(e)

	adminHandler := handler.NewAdminHandler(storageService, episodeService, audioProcessor)
	adminGroup := e.Group("/admin", handler.AuthMiddleware(userRepo))
	adminHandler.RegisterRoutes(adminGroup)

	e.GET("/healthcheck", func(c echo.Context) error {
		if err := db.Ping(c.Request().Context()); err != nil {
			return c.JSON(503, map[string]string{"status": "unhealthy", "error": err.Error()})
		}
		return c.JSON(200, map[string]string{"db": "ok"})
	})

	// Static files — must be registered LAST so it doesn't swallow routes
	e.Static("/", "public")

	log.Printf("starting Castogo on: %s (%s)", config.Cfg.Port, config.Cfg.Env)
	e.Logger.Fatal(e.Start(":" + config.Cfg.Port))
}
