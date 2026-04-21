package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	recoverer "github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/session"
	fiberPostgres "github.com/gofiber/storage/postgres/v3"
	"github.com/gracchi-stdio/castogo/internal/config"
	"github.com/gracchi-stdio/castogo/internal/repository/postgres"
)

func main() {
	err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	app := fiber.New(fiber.Config{
		AppName: "Castogo",
	})

 log.Print(config.Cfg.DatabaseURL)
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



	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Ok")
	})

app.Get("/healthcheck", func (c fiber.Ctx) error {
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


	log.Printf("starting Castogo on: %s (%s)", config.Cfg.Port, config.Cfg.Env)

	if err := app.Listen(":" + config.Cfg.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
