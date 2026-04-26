package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"wildberries-go-course/L0/cache"
	"wildberries-go-course/L0/consumers"
	"wildberries-go-course/L0/database"
	handlers "wildberries-go-course/L0/handler"
	"wildberries-go-course/L0/repository"

	"github.com/joho/godotenv"
)

func main() {
	logLevel := &slog.LevelVar{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)
	logLevel.Set(slog.LevelDebug)
	ctx := context.Background()

	db := database.New()

	if err := db.RunMigration(database.MigrationFolderPath); err != nil {
		log.Fatal("Migration failed with error: ", err)
	}
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	userRep := repository.NewOrderRepository(db)
	orders, err := userRep.GetLastNOrders(ctx, 20)
	if err != nil {
		log.Fatal("Filed to get orders for the cache", err)
	}
	for _, order := range orders {
		cache.GlobalOrderCache.Put(order.OrderUID, order)
	}
	var waitgroup sync.WaitGroup

	mux := http.NewServeMux()

	mux.HandleFunc("/order/", handlers.GetOrderByID(userRep))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	if err != nil {
		log.Fatal("Failed to create user")
	}

	waitgroup.Add(1)
	waitgroup.Go(func() {
		defer waitgroup.Done()
		consumers.RunDeliveryConsumer(userRep)
	})
	waitgroup.Add(1)

	waitgroup.Go(func() {
		defer waitgroup.Done()
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal("Error from the server:", err)
		}
	})

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server shutdown:", slog.Any("error", err))
	}
	waitgroup.Wait()
}
