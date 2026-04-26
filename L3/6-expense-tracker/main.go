package main

import (
	"context"
	"net/http"
	"wildberries-go-course/L3-6/database"
	"wildberries-go-course/L3-6/handler"

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

	router.GET("/items", handler.GetRecords(ctx, db))
	router.POST("/items", handler.CreateRecord(ctx, db))
	router.PUT("/items/:id", handler.UpdateRecord(ctx, db))
	router.DELETE("/items/:id", handler.DeleteRecord(ctx, db))
	router.GET("/analytics", handler.GetAnalytics(ctx, db))

	zlog.Logger.Info().Msg("Starting server on port 8080")
	if err := router.Run(); err != nil && err != http.ErrServerClosed {
		zlog.Logger.Fatal().Msg("Server startup failed: " + err.Error())
	}

	doneCleanerChan <- true

	zlog.Logger.Info().Msg("Server closed")
}
