package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"
	"wildberries-go-course/L0/cache"
	"wildberries-go-course/L0/repository"

	"github.com/google/uuid"
)

func GetOrderByID(orderRep repository.OrderRepositoryInterface) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		start := time.Now()
		if request.Method != http.MethodGet {
			http.Error(responseWriter, "Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		orderUUID := strings.TrimPrefix(request.URL.Path, "/order/")
		err := uuid.Validate(orderUUID)
		if err != nil {
			slog.Debug("Bad UUID, ", "error", err)
			http.Error(responseWriter, "Bad Id", http.StatusBadRequest)
			return
		}

		order, exist := cache.GlobalOrderCache.Get(orderUUID)
		if !exist {
			order, err = orderRep.GetOrderByID(context.Background(), &orderUUID)
			if err != nil {
				http.Error(responseWriter, "Not Found", http.StatusNotFound)
				return
			}
			cache.GlobalOrderCache.Put(orderUUID, order)
		}

		responseWriter.Header().Set("Content-Type", "application/json")
		responseWriter.Header().Set("Access-Control-Allow-Origin", "*")

		json.NewEncoder(responseWriter).Encode(order)
		duration := time.Since(start)
		slog.Debug("Function took ", "duration", duration)
	}
}
