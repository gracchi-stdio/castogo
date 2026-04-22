package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gracchi-stdio/castogo/internal/config"
	"github.com/gracchi-stdio/castogo/internal/repository/postgres"
	"github.com/gracchi-stdio/castogo/internal/service"
	"github.com/urfave/cli/v3"
)

func main() {

	err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := postgres.NewPool(context.Background(), config.Cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := db.Ping(context.Background()); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	// services setup
	userRepo := postgres.NewUserRepo(db)
	authService := service.NewAuthService(userRepo)

	log.Println("Database connection successful, services initialized")

	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "email",
				Usage:    "Email for registration",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "password",
				Usage:    "Password for registration",
				Required: true,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			email := cmd.String("email")
			password := cmd.String("password")

			user, err := authService.Register(ctx, email, password)
			if err != nil {
				return fmt.Errorf("failed to create user: %w", err)
			}
			log.Printf("user %s successfully created", user.Email)
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
