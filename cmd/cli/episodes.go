package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gracchi-stdio/castogo/internal/repository"
	"github.com/urfave/cli/v3"
)

func episodesListCommand() *cli.Command {
	return &cli.Command{
		Name:  "episodes-list",
		Usage: "List episodes (id, status, title)",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "status", Usage: "draft|scheduled|published|archied (default: all)"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			db, err := connectDatabase(ctx, cmd)
			if err != nil {
				return err
			}
			defer db.Close()

			episodes, err := db.EpisodeRepo().List(ctx, repository.EpisodeFilter{
				Status: cmd.String("status"),
				Limit:  20,
			})
			if err != nil {
				return fmt.Errorf("list episodes: %w", err)

			}

			for _, ep := range episodes {
				log.Printf("%d  %-10s  %s", ep.ID, ep.Status(), ep.Title)
			}
			log.Printf("%d episode(s)", len(episodes))
			return nil
		},
	}
}
