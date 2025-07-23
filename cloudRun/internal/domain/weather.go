package domain

// WeatherResponse representa a resposta com informações de temperatura
// @Description Resposta contendo a temperatura em Celsius, Fahrenheit e Kelvin
type WeatherResponse struct {
	TempC float64 `json:"temp_C" example:"28.5" description:"Temperatura em Celsius"`
	TempF float64 `json:"temp_F" example:"83.3" description:"Temperatura em Fahrenheit"`
	TempK float64 `json:"temp_K" example:"301.5" description:"Temperatura em Kelvin"`
}

// ErrorResponse representa uma resposta de erro
// @Description Resposta de erro da API
type ErrorResponse struct {
	Message string `json:"message" example:"invalid zipcode" description:"Mensagem de erro"`
}

// ViaCEPResponse representa a resposta da API ViaCEP
type ViaCEPResponse struct {
	CEP        string `json:"cep"`
	Logradouro string `json:"logradouro"`
	Bairro     string `json:"bairro"`
	Localidade string `json:"localidade"`
	UF         string `json:"uf"`
	Erro       bool   `json:"erro,omitempty"`
}

// WeatherAPIResponse representa a resposta da API de clima
type WeatherAPIResponse struct {
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

// Location representa uma localização
type Location struct {
	City  string
	State string
}
