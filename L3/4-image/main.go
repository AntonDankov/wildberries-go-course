package main

import (
	"context"
	"net/http"
	"wildberries-go-course/L3-4/database"
	"wildberries-go-course/L3-4/handler"
	"wildberries-go-course/L3-4/kafka"
	"wildberries-go-course/L3-4/storage"
	"wildberries-go-course/L3-4/util"

	"github.com/gin-contrib/cors"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

func main() {
	zlog.Init()

	if err := util.InitWatermark("assets/wb-watermark.png"); err != nil {
		panic(err)
	}

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

	imageStorage := storage.ImageStorage{
		BasePath: "images",
		Depth:    2,
		Width:    2,
	}

	ctx := context.Background()

	brokers := []string{"localhost:9092"}
	topic := "image-processing"
	groupID := "image-workers"

	if err := kafka.CreateTopic(brokers, topic, 3, 1); err != nil {
		panic(err)
	}

	imageConsumer := kafka.NewImageConsumer(brokers, topic, groupID, &imageStorage)

	imageProducer := kafka.NewImageProducer(brokers, topic)

	go imageConsumer.Start(ctx, db)

	router.POST("/upload", handler.UploadImage(ctx, db, &imageStorage, imageProducer))
	router.GET("/image/:id", handler.GetImage(ctx, db, &imageStorage))
	router.GET("/image", handler.GetImagesStatus(ctx, db))
	router.DELETE("/image/:id", handler.DeleteImage(ctx, db, &imageStorage))

	zlog.Logger.Info().Msg("Starting server on port 8080")
	if err := router.Run(); err != nil && err != http.ErrServerClosed {
		zlog.Logger.Fatal().Msg("Server startup failed: " + err.Error())
	}

	zlog.Logger.Info().Msg("Server closed")
}
