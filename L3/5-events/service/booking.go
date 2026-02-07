package service

import (
	"context"
	"time"
	"wildberries-go-course/L3-5/database"
	"wildberries-go-course/L3-5/repository"

	"github.com/wb-go/wbf/zlog"
)

func ScheduledBookCleaner(ctx context.Context, db *database.Database, interval time.Duration, done chan bool) {
	zlog.Logger.Info().Msg("Launching scheduled book cleaner")
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			err := repository.CancelTimeoutedBooks(ctx, db.Master)
			if err != nil {
				zlog.Logger.Error().Msgf("%v", err)
			}
		case <-done:
			return
		case <-ctx.Done():
			return
		}
	}
}
