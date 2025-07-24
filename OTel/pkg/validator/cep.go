package validator

import (
	"regexp"
	"strings"
)

// ValidateCEP validates Brazilian postal code format
func ValidateCEP(cep string) bool {
	// Remove traços e espaços
	cep = strings.ReplaceAll(cep, "-", "")
	cep = strings.ReplaceAll(cep, " ", "")

	// Verifica se tem exatamente 8 dígitos
	if len(cep) != 8 {
		return false
	}

	// Verifica se todos são números
	matched, _ := regexp.MatchString(`^\d{8}$`, cep)
	return matched
}

// CleanCEP removes dashes and spaces from CEP
func CleanCEP(cep string) string {
	cep = strings.ReplaceAll(cep, "-", "")
	cep = strings.ReplaceAll(cep, " ", "")
	return cep
}

// FormatCEP formats CEP with dash (XXXXX-XXX)
func FormatCEP(cep string) string {
	// Clean the CEP first
	cleaned := CleanCEP(cep)

	// Add dash if it has 8 digits
	if len(cleaned) == 8 {
		return cleaned[:5] + "-" + cleaned[5:]
	}

	return cleaned
}
