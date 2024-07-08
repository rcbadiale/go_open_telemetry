package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rcbadiale/go_open_telemetry/pkg/logging"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

const (
	ViaCEP_URL = "https://viacep.com.br/ws/%s/json/"
)

type CEPService interface {
	GetAddressByCEP(ctx context.Context, cep string) (*ViaCEPResponse, error)
}

// ViaCEPService is a service to interact with the ViaCEP API
type ViaCEPService struct {
	BaseHttpService
}

type ViaCEPResponse struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
	Erro        bool   `json:"erro,omitempty"`
}

// NewViaCEPService creates a new ViaCEPService
func NewViaCEPService() CEPService {
	return &ViaCEPService{
		BaseHttpService{
			Client: &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)},
			Tracer: otel.Tracer(""),
		},
	}
}

// GetAddressByCEP returns the address for a given CEP
func (v *ViaCEPService) GetAddressByCEP(ctx context.Context, cep string) (*ViaCEPResponse, error) {
	otel.SetTextMapPropagator(propagation.TraceContext{})
	ctx, span := v.Tracer.Start(ctx, "ViaCEPService.GetAddressByCEP")
	defer span.End()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(ViaCEP_URL, cep), nil)
	if err != nil {
		logging.Logger.Error("Error generating CEP request: ", err)
		return nil, err
	}

	span.AddEvent("Launching Request to external service")
	resp, err := v.Client.Do(req)
	if err != nil {
		logging.Logger.Error("Error getting address by CEP: ", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logging.Logger.Error("Error reading response body: ", err)
		return nil, err
	} else if resp.StatusCode != 200 {
		return nil, ErrInvalidCEP
	}

	var viaCepResponse ViaCEPResponse
	err = json.Unmarshal(body, &viaCepResponse)
	if err != nil {
		logging.Logger.Error("Error unmarshalling response body: ", err)
		return nil, err
	} else if viaCepResponse.Erro {
		logging.Logger.Error(fmt.Sprintf("Error invalid address by CEP: %v", viaCepResponse))
		return nil, ErrCEPNotFound
	}

	return &viaCepResponse, nil
}
