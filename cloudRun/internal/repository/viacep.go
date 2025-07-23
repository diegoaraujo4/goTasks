package repository

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"cloudrun/internal/domain"
)

// ViaCEPRepository handles communication with ViaCEP API
type ViaCEPRepository struct {
	client  *http.Client
	baseURL string
}

// NewViaCEPRepository creates a new ViaCEP repository
func NewViaCEPRepository() *ViaCEPRepository {
	return &ViaCEPRepository{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: "https://viacep.com.br/ws",
	}
}

// GetLocationByCEP fetches location data from ViaCEP API
func (r *ViaCEPRepository) GetLocationByCEP(cep string) (*domain.ViaCEPResponse, error) {
	url := fmt.Sprintf("%s/%s/json/", r.baseURL, cep)

	resp, err := r.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch location data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ViaCEP API returned status %d", resp.StatusCode)
	}

	var viacepResp domain.ViaCEPResponse
	if err := json.NewDecoder(resp.Body).Decode(&viacepResp); err != nil {
		return nil, fmt.Errorf("failed to decode ViaCEP response: %w", err)
	}

	if viacepResp.Erro {
		return nil, fmt.Errorf("CEP not found")
	}

	return &viacepResp, nil
}
