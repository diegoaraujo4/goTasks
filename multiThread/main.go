package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type BrasilAPIResponse struct {
	CEP      string `json:"cep"`
	State    string `json:"state"`
	City     string `json:"city"`
	District string `json:"district"`
	Street   string `json:"street"`
	Service  string `json:"service"`
	Location struct {
		Type        string    `json:"type"`
		Coordinates []float64 `json:"coordinates"`
	} `json:"location"`
}

type ViaCEPResponse struct {
	CEP         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	UF          string `json:"uf"`
	IBGE        string `json:"ibge"`
	GIA         string `json:"gia"`
	DDD         string `json:"ddd"`
	SIAFI       string `json:"siafi"`
}

type CEPResult struct {
	CEP      string
	Street   string
	District string
	City     string
	State    string
	Source   string
}

func fetchBrasilAPI(cep string, ch chan<- CEPResult) {
	url := fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep)

	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var result BrasilAPIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return
	}

	ch <- CEPResult{
		CEP:      result.CEP,
		Street:   result.Street,
		District: result.District,
		City:     result.City,
		State:    result.State,
		Source:   "BrasilAPI",
	}
}

func fetchViaCEP(cep string, ch chan<- CEPResult) {
	url := fmt.Sprintf("http://viacep.com.br/ws/%s/json/", cep)

	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var result ViaCEPResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return
	}

	ch <- CEPResult{
		CEP:      result.CEP,
		Street:   result.Logradouro,
		District: result.Bairro,
		City:     result.Localidade,
		State:    result.UF,
		Source:   "ViaCEP",
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run main.go <CEP>")
		fmt.Println("Exemplo: go run main.go 01153000")
		os.Exit(1)
	}

	cep := os.Args[1]

	if len(cep) != 8 {
		fmt.Println("Erro: CEP deve ter 8 dígitos")
		fmt.Println("Exemplo: 01153000")
		os.Exit(1)
	}

	ch := make(chan CEPResult, 2)

	fmt.Printf("🔍 Buscando CEP %s nas APIs BrasilAPI e ViaCEP...\n", cep)
	start := time.Now()

	go fetchBrasilAPI(cep, ch)
	go fetchViaCEP(cep, ch)

	select {
	case result := <-ch:
		elapsed := time.Since(start)
		fmt.Printf("\n✅ === RESULTADO MAIS RÁPIDO ===\n")
		fmt.Printf("🏆 API Vencedora: %s\n", result.Source)
		fmt.Printf("📮 CEP: %s\n", result.CEP)
		fmt.Printf("🏠 Logradouro: %s\n", result.Street)
		fmt.Printf("🏘️  Bairro: %s\n", result.District)
		fmt.Printf("🏙️  Cidade: %s\n", result.City)
		fmt.Printf("🗺️  Estado: %s\n", result.State)
		fmt.Printf("⏱️  Tempo de resposta: %v\n", elapsed.Round(time.Millisecond))

	case <-time.After(1 * time.Second):
		fmt.Println("\n❌ Erro: Timeout - Nenhuma API respondeu em 1 segundo")
		os.Exit(1)
	}
}
