package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rcbadiale/go_open_telemetry/internals/services"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type GetWeatherResponse struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

type WeatherHandler struct {
	CEPService     services.CEPService
	WeatherService services.WeatherService
	Tracer         trace.Tracer
}

func NewWeatherHandler(weatherApiKey string, tracer trace.Tracer) *WeatherHandler {
	return &WeatherHandler{
		Tracer:         tracer,
		CEPService:     services.NewViaCEPService(),
		WeatherService: services.NewWeatherAPIService(weatherApiKey),
	}
}

// GetWeather returns the weather
func (wh *WeatherHandler) GetWeather(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	carrier := propagation.HeaderCarrier(r.Header)
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)

	w.Header().Set("Content-Type", "application/json")
	zipCode := chi.URLParam(r, "zipCode")
	if len(zipCode) != 8 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte("invalid zipcode"))
		return
	}
	responseCEP, error := wh.CEPService.GetAddressByCEP(ctx, zipCode)
	if error != nil {
		switch error {
		case services.ErrCEPNotFound:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("can not find zipcode"))
			return
		case services.ErrInvalidCEP:
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("invalid zipcode"))
			return
		default:
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal server error"))
			return
		}
	}
	responseWeather, error := wh.WeatherService.GetWeatherByCity(ctx, responseCEP.Localidade)
	if error != nil {
		switch error {
		case services.ErrCEPNotFound:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("can not find zipcode"))
			return
		case services.ErrInvalidCEP:
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("invalid zipcode"))
			return
		default:
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal server error"))
			return
		}
	}
	// Truncating values to ensure only 1 decimal place
	output := GetWeatherResponse{
		City:  responseCEP.Localidade,
		TempC: float64(int(responseWeather.Current.TempC*10)) / 10,
		TempF: float64(int(responseWeather.Current.TempF*10)) / 10,
		TempK: float64(int((responseWeather.Current.TempC+273.15)*10)) / 10,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(output)
}
