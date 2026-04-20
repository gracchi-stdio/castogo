package main

import (
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/gracchi-stdio/podlog/internal/config"
)

func main() {
	err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	app := fiber.New(fiber.Config{
		AppName: "Padlog",
	})

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Ok")
	})

	log.Printf("starting podlog on: %s (%s)", config.Cfg.Port, config.Cfg.Env)

	if err := app.Listen(":" + config.Cfg.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
