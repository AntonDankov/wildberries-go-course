package main

import (
	"context"
	"net/http"
	"wildberries-go-course/L3-7/database"
	"wildberries-go-course/L3-7/handler"

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

	// user
	router.POST("/user", handler.Register(ctx, db))
	router.POST("/login", handler.Login(ctx, db))
	// items
	router.GET("/items", handler.GetItems(ctx, db))
	router.GET("/items/:id", handler.GetItem(ctx, db))
	router.POST("/items", handler.CreateItem(ctx, db))
	router.PUT("/items/:id", handler.UpdateItem(ctx, db))
	router.DELETE("/items/:id", handler.DeleteItem(ctx, db))
	router.GET("/analytics/:id", handler.GetItemHistory(ctx, db))

	zlog.Logger.Info().Msg("Starting server on port 8080")
	if err := router.Run(); err != nil && err != http.ErrServerClosed {
		zlog.Logger.Fatal().Msg("Server startup failed: " + err.Error())
	}

	zlog.Logger.Info().Msg("Server closed")
}
