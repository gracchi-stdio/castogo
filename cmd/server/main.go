package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	recoverer "github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/gofiber/fiber/v3/middleware/static"
	fiberPostgres "github.com/gofiber/storage/postgres/v3"
	"github.com/gracchi-stdio/castogo/internal/config"
	"github.com/gracchi-stdio/castogo/internal/handler"
	"github.com/gracchi-stdio/castogo/internal/repository/postgres"
	"github.com/gracchi-stdio/castogo/internal/service"
)

func main() {
	err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	app := fiber.New(fiber.Config{
		AppName: "Castogo",
	})

	app.Use(logger.New())
	app.Use(recoverer.New())

	db, err := postgres.NewPool(context.Background(), config.Cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	pgStorage := fiberPostgres.New(fiberPostgres.Config{
		DB: db,
	})

	app.Use(session.New(session.Config{
		Storage: pgStorage,
	}))

	// Register handlers and services
	userRepo := postgres.NewUserRepo(db)
	authService := service.NewAuthService(userRepo)
	authHandler := handler.NewAuthHandler(authService)
	authHandler.RegisterRoutes(app)

	clockHandler := handler.NewClockHandler()
	clockHandler.RegisterRoutes(app)

	filterHandler := handler.NewFilterHandler()
	filterHandler.RegisterRoutes(app)

	app.Get("/healthcheck", func(c fiber.Ctx) error {
		if err := db.Ping(c.Context()); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(
				fiber.Map{
					"status": "unhealthy",
					"error":  err.Error(),
				},
			)
		}
		return c.JSON(fiber.Map{"db": "ok"})
	})

	// Static files — must be registered LAST so it doesn't swallow routes
	app.Use("/*", static.New("./public"))

	log.Printf("starting Castogo on: %s (%s)", config.Cfg.Port, config.Cfg.Env)

	if err := app.Listen(":" + config.Cfg.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
