package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rcbadiale/go_open_telemetry/internals/handlers"
	"github.com/rcbadiale/go_open_telemetry/internals/server"
	"github.com/rcbadiale/go_open_telemetry/pkg/environment"
	"github.com/rcbadiale/go_open_telemetry/pkg/logging"
	"github.com/rcbadiale/go_open_telemetry/pkg/telemetry"
	"github.com/riandyrn/otelchi"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	logger := logging.SetupLogger()
	logging.Logger.Info("Starting input service at :8080")
	if err := run(logger); err != nil {
		logging.Logger.Error("Failed to run the server", err)
		panic(err)
	}
}

func run(logger *log.Logger) (err error) {
	serviceName := environment.GetEnvOrDefault("SERVICE_NAME", "input-service")
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	otelEndpoint := environment.GetEnvOrDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317")
	shutdown, err := telemetry.InitProvider(ctx, serviceName, otelEndpoint)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Fatal("failed to shutdown TraceProvider", err)
		}
	}()

	tracer := otel.Tracer(serviceName)
	r := setupHandler(tracer, serviceName)

	srv := server.StartServer(ctx, r, ":8080", logger)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logging.Logger.Error("failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	select {
	case <-sigCh:
		logging.Logger.Warn("Shutting down gracefully, CTRL+C pressed...")
	case <-ctx.Done():
		logging.Logger.Warn("Shutting down due to other reason...")
	}

	// Create a timeout context for the graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logging.Logger.Warn("Server forced to shutdown: %v", err)
	}

	logging.Logger.Warn("Server exiting")
	return nil
}

func setupHandler(trace trace.Tracer, serviceName string) *chi.Mux {
	r := chi.NewRouter()
	r.Use(otelchi.Middleware(serviceName, otelchi.WithChiRoutes(r)))
	weatherHandler := handlers.NewOtelWeatherInputHandler(trace)
	r.Post("/weather", weatherHandler.PostWeather)
	// r.Handle("/metrics", promhttp.Handler())
	return r
}
