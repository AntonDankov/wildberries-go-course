package main

import (
	"context"
	"net/http"
	"time"
	"wildberries-go-course/L3-5/database"
	"wildberries-go-course/L3-5/handler"
	"wildberries-go-course/L3-5/service"

	"github.com/gin-contrib/cors"
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

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))

	ctx := context.Background()

	doneCleanerChan := make(chan bool)

	go service.ScheduledBookCleaner(ctx, db, time.Minute, doneCleanerChan)

	router.POST("/events", handler.CreateEvent(ctx, db))
	router.GET("/events", handler.GetEvents(ctx, db))
	router.GET("/events/:id", handler.GetEvent(ctx, db))
	router.GET("/events/:id/book", handler.GetEventBooks(ctx, db))
	router.POST("/events/:id/book", handler.BookEvent(ctx, db))
	router.POST("/events/:id/confirm", handler.ConfirmBook(ctx, db))

	zlog.Logger.Info().Msg("Starting server on port 8080")
	if err := router.Run(); err != nil && err != http.ErrServerClosed {
		zlog.Logger.Fatal().Msg("Server startup failed: " + err.Error())
	}

	doneCleanerChan <- true

	zlog.Logger.Info().Msg("Server closed")
}
