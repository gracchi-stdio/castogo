package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"

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

			// Step 4: db grows UserRepo() when the libsql repositories land,
			// then register becomes dialect-blind like every other command:
			//
			//   db, err := connectDatabase(ctx, cmd)
			//   authService := service.NewAuthService(db.UserRepo())
			//   user, err := authService.Register(ctx, email, password)
			_ = email
			_ = password
			_ = generated
			return fmt.Errorf("register: arrives with the libsql user repository (step 4)")
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
