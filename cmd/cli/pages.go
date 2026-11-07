package main

import (
	"context"
	"fmt"
	"log"

	"github.com/urfave/cli/v3"
)

func pagesListCommand() *cli.Command {
	return &cli.Command{
		Name:  "pages-list",
		Usage: "List pages (id, nav, published, path)",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			db, err := connectDatabase(ctx, cmd)
			if err != nil {
				return err
			}
			defer db.Close()

			pages, err := db.PageRepo().List(ctx)
			if err != nil {
				return fmt.Errorf("list pages: %w", err)
			}
			for _, p := range pages {
				nav := "hidden"
				if p.NavOrder != nil {
					nav = fmt.Sprintf("order %d", *p.NavOrder)
				}
				state := "draft"
				if p.IsPublished {
					state = "published"
				}
				log.Printf("%d %-9s %-9s %s", p.ID, nav, state, p.Path)
			}
			log.Printf("%d page(s)", len(pages))
			return nil
		},
	}
}
