package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gracchi-stdio/castogo/internal/version"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "castogo-cli",
		Usage: "CLI utilities for Castogo",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "database-url",
				Usage: "Database URL for this invocation — overrides DATABASE_URL (env/.env). Used to target another environment, e.g. production from a dev machine.",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "version",
				Usage: "Print the Castogo build version",
				Action: func(context.Context, *cli.Command) error {
					fmt.Println(version.String())
					return nil
				},
			},
			registerCommand(),
			migrateResetCommand(),
			migrateCommand(),
			migrateAdoptCommand(),
			dbPingCommand(),
			episodesListCommand(),
			pagesListCommand(),
			analyticsSmokeCommand(),
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
