package server

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func StartServer(ctx context.Context, r *chi.Mux, address string, logger *log.Logger) *http.Server {
	srv := &http.Server{
		Addr:         address,
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      r,
		ErrorLog:     logger,
	}
	return srv
}
