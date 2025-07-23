package domain

// WeatherService define a interface para serviços de clima
type WeatherService interface {
	GetLocationByCEP(cep string) (*ViaCEPResponse, error)
	GetWeatherByLocation(location string) (*WeatherAPIResponse, error)
}

// LocationService define a interface para serviços de localização
type LocationService interface {
	GetLocationByCEP(cep string) (*ViaCEPResponse, error)
}

// WeatherDataService define a interface para dados meteorológicos
type WeatherDataService interface {
	GetWeatherByLocation(location string) (*WeatherAPIResponse, error)
}
