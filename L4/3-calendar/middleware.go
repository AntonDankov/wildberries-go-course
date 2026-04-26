package main

import (
	"log"
	"net/http"
	"time"
)

type Middleware func(next http.Handler) http.Handler

func Chain(middlewares ...Middleware) http.Handler {
	var handler http.Handler

	N := len(middlewares)
	for i := range middlewares {
		middlewares[N-1-i](handler)
	}

	return handler
}

func Log(nextHandler http.Handler) http.Handler {
	return http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			nextHandler.ServeHTTP(writer, request)

			log.Printf("[%s] %s %v", request.Method, request.URL.Path, time.Now())
		})
}
