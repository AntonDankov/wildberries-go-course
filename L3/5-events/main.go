package main

import (
	"context"
	"net/http"
	"time"
	"wildberries-go-course/L3-5/database"
	"wildberries-go-course/L3-5/handler"
	"wildberries-go-course/L3-5/service"

	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

func main() {
	zlog.Init()

	// Database setup
	db := database.New()

	if err := db.RunMigration(database.MigrationFolderPath); err != nil {
		panic(err)
	}

	// Web server setup
	router := ginext.New("")

	ctx := context.Background()

	doneCleanerChan := make(chan bool)

	go service.ScheduledBookCleaner(ctx, db, time.Minute, doneCleanerChan)

	router.POST("/events", handler.CreateEvent(ctx, db))
	router.POST("/events/:id/book", handler.BookEvent(ctx, db))
	router.POST("/events/:id/confirm", handler.ConfirmBook(ctx, db))
	router.GET("/events/:id", handler.GetEvent(ctx, db))

	zlog.Logger.Info().Msg("Starting server on port 8080")
	if err := router.Run(); err != nil && err != http.ErrServerClosed {
		zlog.Logger.Fatal().Msg("Server startup failed: " + err.Error())
	}

	doneCleanerChan <- true

	zlog.Logger.Info().Msg("Server closed")
}
