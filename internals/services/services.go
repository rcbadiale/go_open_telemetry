package services

import (
	"github.com/rcbadiale/go_open_telemetry/internals"
	"go.opentelemetry.io/otel/trace"
)

type BaseHttpService struct {
	Client internals.HTTPClient
	Tracer trace.Tracer
}
