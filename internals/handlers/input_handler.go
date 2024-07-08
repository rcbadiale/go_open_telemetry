package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/rcbadiale/go_open_telemetry/internals/services"
	"github.com/rcbadiale/go_open_telemetry/pkg/logging"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type PostOtelWeatherResponse struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

type PostOtelWeatherInput struct {
	Cep string `json:"cep"`
}

type OtelWeatherInputHandler struct {
	InternalService services.InternalWeatherService
	OTELTracer      trace.Tracer
}

func NewOtelWeatherInputHandler(trace trace.Tracer) *OtelWeatherInputHandler {
	return &OtelWeatherInputHandler{
		InternalService: services.NewInternalWeatherService(),
		OTELTracer:      trace,
	}
}

// GetWeather returns the weather
func (wh *OtelWeatherInputHandler) PostWeather(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	carrier := propagation.HeaderCarrier(r.Header)
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)

	_, span := wh.OTELTracer.Start(ctx, "OtelWeatherInputHandler.PostWeather")
	defer span.End()

	var input PostOtelWeatherInput
	err := json.NewDecoder(r.Body).Decode(&input)

	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte("invalid zipcode"))
		return
	}
	zipCode := input.Cep
	if len(zipCode) != 8 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte("invalid zipcode"))
		return
	}
	response, error := wh.InternalService.GetWeather(ctx, zipCode)
	if error != nil {
		var statusCode int
		var message string
		switch error {
		case services.ErrCEPNotFound:
			statusCode = http.StatusNotFound
			message = "can not find zipcode"
		case services.ErrInvalidCEP:
			statusCode = http.StatusUnprocessableEntity
			message = "invalid zipcode"
		default:
			statusCode = http.StatusInternalServerError
			message = "internal server error"
		}
		logging.Logger.ErrorContext(ctx, message, error)
		w.WriteHeader(statusCode)
		w.Write([]byte(message))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
