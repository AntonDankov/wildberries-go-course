package main

import (
	"net/http"
	_ "net/http/pprof"
	"wildberries-go-course/L4-4/entity"
	"wildberries-go-course/L4-4/handler"

	"github.com/gin-contrib/cors"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

func main() {
	zlog.Init()

	// Web server setup
	router := ginext.New("")
	entity.SetGCPercentage(67)

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))

	router.GET("/memory_metrics", handler.GetMemoryMetrics())
	router.POST("/gc_percentage", handler.ChangeGarbageCollectorPercentage())

	zlog.Logger.Info().Msg("Starting server on port 8080")
	if err := router.Run(); err != nil && err != http.ErrServerClosed {
		zlog.Logger.Fatal().Msg("Server startup failed: " + err.Error())
	}

	zlog.Logger.Info().Msg("Server closed")
}
