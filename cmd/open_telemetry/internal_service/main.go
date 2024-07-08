package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/rcbadiale/go_open_telemetry/internals/handlers"
	"github.com/rcbadiale/go_open_telemetry/internals/server"
	"github.com/rcbadiale/go_open_telemetry/pkg/logging"
	"github.com/rcbadiale/go_open_telemetry/pkg/telemetry"
	"github.com/riandyrn/otelchi"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	logger := logging.SetupLogger()
	err := godotenv.Load()
	if err != nil {
		logging.Logger.Error("error loading .env file, will use environment variables")
	}

	logging.Logger.Info("Starting server on port 8081")
	if err := run(logger); err != nil {
		logging.Logger.Error("Failed to run the server", err)
		panic(err)
	}
}

const serviceName = "weather-service"

func run(logger *log.Logger) (err error) {
	// Graceful shutdown - begin
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	// Graceful shutdown - end
	otelEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otelEndpoint == "" {
		otelEndpoint = "localhost:4317"
	}
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

	r := setupHandler(tracer)

	srv := server.StartServer(ctx, r, ":8081", logger)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	select {
	case <-sigCh:
		log.Println("Shutting down gracefully, CTRL+C pressed...")
	case <-ctx.Done():
		log.Println("Shutting down due to other reason...")
	}

	// Create a timeout context for the graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
	return nil
}

func setupHandler(trace trace.Tracer) *chi.Mux {
	r := chi.NewRouter()
	r.Use(otelchi.Middleware(serviceName, otelchi.WithChiRoutes(r)))
	weatherApiKey := os.Getenv("WEATHER_API_KEY")
	weatherHandler := handlers.NewWeatherHandler(weatherApiKey, trace)
	r.Get("/weather/{zipCode}", weatherHandler.GetWeather)
	return r
}
