package temperature

import "testing"

func TestConvertCelsiusToFahrenheit(t *testing.T) {
	tests := []struct {
		name     string
		celsius  float64
		expected float64
	}{
		{"Zero celsius", 0, 32},
		{"Room temperature", 20, 68},
		{"Body temperature", 37, 98.6},
		{"Boiling point", 100, 212},
		{"Negative temperature", -10, 14},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertCelsiusToFahrenheit(tt.celsius)
			// Use tolerance for floating point comparison
			if diff := result - tt.expected; diff < -0.01 || diff > 0.01 {
				t.Errorf("ConvertCelsiusToFahrenheit(%v) = %v, want %v", tt.celsius, result, tt.expected)
			}
		})
	}
}

func TestConvertCelsiusToKelvin(t *testing.T) {
	tests := []struct {
		name     string
		celsius  float64
		expected float64
	}{
		{"Zero celsius", 0, 273},
		{"Room temperature", 20, 293},
		{"Body temperature", 37, 310},
		{"Boiling point", 100, 373},
		{"Absolute zero", -273, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertCelsiusToKelvin(tt.celsius)
			if result != tt.expected {
				t.Errorf("ConvertCelsiusToKelvin(%v) = %v, want %v", tt.celsius, result, tt.expected)
			}
		})
	}
}
