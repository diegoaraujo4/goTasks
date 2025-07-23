package temperature

// ConvertCelsiusToFahrenheit converts Celsius to Fahrenheit
func ConvertCelsiusToFahrenheit(celsius float64) float64 {
	return celsius*1.8 + 32
}

// ConvertCelsiusToKelvin converts Celsius to Kelvin
func ConvertCelsiusToKelvin(celsius float64) float64 {
	return celsius + 273
}

// ConvertFahrenheitToCelsius converts Fahrenheit to Celsius
func ConvertFahrenheitToCelsius(fahrenheit float64) float64 {
	return (fahrenheit - 32) / 1.8
}

// ConvertKelvinToCelsius converts Kelvin to Celsius
func ConvertKelvinToCelsius(kelvin float64) float64 {
	return kelvin - 273
}
