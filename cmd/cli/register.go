package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"

	"github.com/gracchi-stdio/castogo/internal/service"
	"github.com/urfave/cli/v3"
)

func registerCommand() *cli.Command {
	return &cli.Command{
		Name:  "register",
		Usage: "Register a new user",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "email",
				Usage:    "Email for registration",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "password",
				Usage: "Password (generated and printed once when omitted)",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			email := cmd.String("email")
			password := cmd.String("password")

			generated := password == ""
			if generated {
				var err error
				password, err = generatePassword(20)
				if err != nil {
					return fmt.Errorf("generate password: %w", err)
				}
			}

			db, err := connectDatabase(ctx, cmd)
			if err != nil {
				return err
			}

			authService := service.NewAuthService(db.UserRepo())

			user, err := authService.Register(ctx, email, password)
			if err != nil {
				return fmt.Errorf("register: %w", err)
			}

			log.Printf("user registered: %s", user.Email)
			if generated {
				log.Printf("generated password (shown once): %s", password)
			}

			return nil
		},
	}
}

// generatePassword returns a cryptographically random password drawn from an
// unambiguous charset (no 0/O/1/l/I) so it is safe to transcribe by hand.
func generatePassword(n int) (string, error) {
	const charset = "abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789!@#_-+"
	buf := make([]byte, n)
	for i := range buf {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		buf[i] = charset[idx.Int64()]
	}
	return string(buf), nil
}
