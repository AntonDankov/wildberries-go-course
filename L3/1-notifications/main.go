package main

import (
	"net/http"
	"os"

	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"

	broker "widlberries-go-course/L3-1/broker"
	database "widlberries-go-course/L3-1/database"
	handler "widlberries-go-course/L3-1/handler"
	repository "widlberries-go-course/L3-1/repository"
)

func main() {
	zlog.Init()

	// Database etup
	db := database.New()

	db.RunMigration(database.MigrationFolderPath)

	pgrep := repository.NewPostgresRepository(db)

	// RabbitMQ setup
	broker.InitRabbitMQ()

	for range os.Getenv("RABBITMQ_CONSUMER_AMOUNT") {
		go func() {
			err := broker.StartConsumer(broker.GlobalRabbitClient, pgrep)
			if err != nil {
				panic(err)
			}
		}()
	}

	// Web server setup
	router := ginext.New("")

	router.POST("/notify", handler.CreateNotificationHandler(pgrep, broker.GlobalRabbitClient))

	router.GET("/notify/:id", handler.GetNotificationHandler(pgrep))
	router.DELETE("/notify/:id", handler.DeleteNotificationHandler(pgrep))

	zlog.Logger.Info().Msg("Starting server on port 8080")
	if err := router.Run(); err != nil && err != http.ErrServerClosed {
		zlog.Logger.Fatal().Msg("Server startup failed: " + err.Error())
	}

	zlog.Logger.Info().Msg("Server closed")
}
