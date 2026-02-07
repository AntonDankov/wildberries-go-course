package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"

	database "wildberries-go-course/L3-2/database"
	handler "wildberries-go-course/L3-2/handler"
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

	router.POST("/shorten", handler.CreateShortLink(ctx, db))
	router.GET("/s/:shortURL", handler.VisitShortLink(ctx, db))

	router.GET("/analytics/:shortURL", func(c *gin.Context) {
		aggregationType := c.Query("type")

		switch aggregationType {
		case "day":
			handler.GetAnalyticsByDay(ctx, db)(c)
		case "month":
			handler.GetAnalyticsByMonth(ctx, db)(c)
		case "user-agent":
			handler.GetAnalyticsByUserAgent(ctx, db)(c)
		default:
			handler.GetAnalyticsFull(ctx, db)(c)
		}
	})

	zlog.Logger.Info().Msg("Starting server on port 8080")
	if err := router.Run(); err != nil && err != http.ErrServerClosed {
		zlog.Logger.Fatal().Msg("Server startup failed: " + err.Error())
	}

	zlog.Logger.Info().Msg("Server closed")
}
