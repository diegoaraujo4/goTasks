package validator

import "testing"

func TestValidateCEP(t *testing.T) {
	tests := []struct {
		name     string
		cep      string
		expected bool
	}{
		{"Valid CEP without dash", "01310100", true},
		{"Valid CEP with dash", "01310-100", true},
		{"Valid CEP with spaces", "01310 100", true},
		{"Invalid CEP too short", "0131010", false},
		{"Invalid CEP too long", "013101000", false},
		{"Invalid CEP with letters", "0131010A", false},
		{"Invalid CEP empty", "", false},
		{"Invalid CEP special chars", "01310@100", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateCEP(tt.cep)
			if result != tt.expected {
				t.Errorf("ValidateCEP(%q) = %v, want %v", tt.cep, result, tt.expected)
			}
		})
	}
}

func TestCleanCEP(t *testing.T) {
	tests := []struct {
		name     string
		cep      string
		expected string
	}{
		{"CEP with dash", "01310-100", "01310100"},
		{"CEP with spaces", "01310 100", "01310100"},
		{"CEP with dash and spaces", "01310- 100", "01310100"},
		{"Clean CEP", "01310100", "01310100"},
		{"Multiple dashes and spaces", "0-1 3-1 0-1 0-0", "01310100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanCEP(tt.cep)
			if result != tt.expected {
				t.Errorf("CleanCEP(%q) = %q, want %q", tt.cep, result, tt.expected)
			}
		})
	}
}
