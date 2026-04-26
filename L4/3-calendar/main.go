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
)

func main() {
	flags := GetFlags()

	eventStorage := NewEventStorage()

	ctx := context.Background()

	go func() {
		runCleanScheduler(ctx, 1*time.Minute, 1*time.Minute, eventStorage)
	}()
	notificationMessageChan := make(chan NotificationEventMessage)
	runNotificationEvents(ctx, notificationMessageChan)

	eventHTTPHandler := NewEventHTTPHandler(eventStorage, notificationMessageChan)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /create_event", eventHTTPHandler.CreateEvent)
	mux.HandleFunc("POST /update_event", eventHTTPHandler.UpdateEvent)
	mux.HandleFunc("POST /delete_event", eventHTTPHandler.DeleteEvent)
	mux.HandleFunc("GET /events_for_day", eventHTTPHandler.GetEventByDate)
	mux.HandleFunc("GET /events_for_week", eventHTTPHandler.GetEventByWeek)
	mux.HandleFunc("GET /events_for_month", eventHTTPHandler.GetEventByMonth)

	loggedMux := Log(mux)

	server := &http.Server{
		Addr:         ":" + flags.Port,
		Handler:      loggedMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,

		IdleTimeout: 2 * time.Minute,
	}

	var waitgroup sync.WaitGroup
	waitgroup.Go(func() {
		log.Printf("Server starting on port %s", flags.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	})

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server shutdown:", slog.Any("error", err))
	}
	waitgroup.Wait()
	log.Printf("Server closed\n")
}
