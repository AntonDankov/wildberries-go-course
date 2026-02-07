package main

import (
	"context"
	"net/http"
	"wildberries-go-course/L3-3/database"
	"wildberries-go-course/L3-3/handler"

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

	router.POST("/comment", handler.CreateComment(ctx, db))
	router.GET("/comment/:commentID", handler.GetCommentWithReplies(ctx, db))
	router.DELETE("/comment/:commentID", handler.DeleteCommentWithReplies(ctx, db))
	router.GET("/comment/search", handler.FindCommentsByText(ctx, db))
	zlog.Logger.Info().Msg("Starting server on port 8080")
	if err := router.Run(); err != nil && err != http.ErrServerClosed {
		zlog.Logger.Fatal().Msg("Server startup failed: " + err.Error())
	}

	zlog.Logger.Info().Msg("Server closed")
}
