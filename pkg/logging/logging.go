package logging

import (
	"log"
	"log/slog"
	"os"
)

var Logger *slog.Logger

func SetupLogger() *log.Logger {
	loggerHandler := slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{
			AddSource: true,
		},
	)
	Logger = slog.New(loggerHandler)
	return slog.NewLogLogger(loggerHandler, slog.LevelInfo)
}
