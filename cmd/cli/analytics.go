package main

import (
	"context"
	"log"
	"time"

	"github.com/urfave/cli/v3"
)

func analyticsSmokeCommand() *cli.Command {
	return &cli.Command{
		Name:  "analytics-smoke",
		Usage: "Round-trip every AnalyticsRepo method",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			db, err := connectDatabase(ctx, cmd)
			if err != nil {
				return err
			}
			defer db.Close()

			repo := db.AnalyticsRepo()
			today := time.Now().UTC()

			if err := repo.InsertProcessedFile(ctx, "/tmp/smoke.log", today, 42); err != nil {
				return err
			}
			seen, err := repo.IsFileProcessed(ctx, "/tmp/smoke.log")
			if err != nil {
				return err
			}
			fresh, err := repo.IsFileProcessed(ctx, "/tmp/never-seen.log")
			if err != nil {
				return err
			}
			log.Printf("IsFileProcessed: known=%v unknown=%v (want true false)", seen, fresh)
			hash := "deadbeef" // 40-char SHA1 hex in prod; this is just a smoke test
			if _, _, err := repo.AccumulateBytes(ctx, hash, 1, today, 100); err != nil {
				return err
			}
			bytesSeen, counted, err := repo.AccumulateBytes(ctx, hash, 1, today, 150)
			if err != nil {
				return err
			}
			log.Printf("AccumulateBytes: bytes=%d counted=%v (want 250 false)", bytesSeen, counted)

			if err := repo.MarkCounted(ctx, hash); err != nil {
				return err
			}
			_, counted, err = repo.AccumulateBytes(ctx, hash, 1, today, 1)
			if err != nil {
				return err
			}
			log.Printf("after MarkCounted: counted=%v (want true)", counted)

			if err := repo.UpsertPodcastDaily(ctx, 1, today, 1024); err != nil {
				return err
			}
			if err := repo.UpsertPodcastDaily(ctx, 1, today, 1024); err != nil {
				return err
			}
			if err := repo.UpsertByHour(ctx, 1, today, 14); err != nil {
				return err
			}
			if err := repo.UpsertByEpisode(ctx, 1, 1, today, 3); err != nil {
				return err
			}
			if err := repo.UpsertByPlayer(ctx, 1, today, "spotify", "", "", "", false); err != nil {
				return err
			}
			if err := repo.UpsertByCountry(ctx, 1, today, "DE"); err != nil {
				return err
			}
			log.Println("analytics-smoke OK")
			return nil

		},
	}
}
