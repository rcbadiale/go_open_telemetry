package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rcbadiale/go_open_telemetry/pkg/environment"
	"github.com/rcbadiale/go_open_telemetry/pkg/logging"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type InternalWeatherResponse struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

type InternalWeatherService interface {
	GetWeather(ctx context.Context, zipCode string) (*InternalWeatherResponse, error)
}

type InternalWeatherAPIService struct {
	BaseHttpService
	ServiceUrl string
}

func (i *InternalWeatherAPIService) GetWeather(ctx context.Context, zipCode string) (*InternalWeatherResponse, error) {
	otel.SetTextMapPropagator(propagation.TraceContext{})
	ctx, span := i.Tracer.Start(ctx, "InternalWeatherAPIService.GetWeather")
	defer span.End()
	if len(zipCode) != 8 {
		return nil, ErrInvalidCEP
	}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		i.ServiceUrl+"/weather/"+zipCode,
		nil,
	)
	if err != nil {
		logging.Logger.Error("Error getting address by CEP: ", err)
		return nil, err
	}
	span.AddEvent("Launching Request to external service")
	resp, err := i.Client.Do(req)
	if err != nil {
		logging.Logger.Error("Error getting weather service: ", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logging.Logger.Error("Error reading response body: ", err)
		return nil, err
	} else if resp.StatusCode != 200 {
		switch resp.StatusCode {
		case 404:
			return nil, ErrCEPNotFound
		default:
			return nil, fmt.Errorf("error getting address from internal service: %v", body)
		}
	}

	var internalWeatherResponse InternalWeatherResponse
	err = json.Unmarshal(body, &internalWeatherResponse)
	if err != nil {
		logging.Logger.Error("Error unmarshalling response body: ", err)
		return nil, err
	}

	return &internalWeatherResponse, nil
}

func NewInternalWeatherService() InternalWeatherService {
	return &InternalWeatherAPIService{
		ServiceUrl: environment.GetEnvOrDefault("WEATHER_SERVICE_URL", "http://localhost:8081"),
		BaseHttpService: BaseHttpService{
			Client: &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)},
			Tracer: otel.Tracer(""),
		},
	}
}
