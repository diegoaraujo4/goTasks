package gateway

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGatewayHandler_ProcessCEP_ValidCEP(t *testing.T) {
	// Create a mock orchestration service
	mockOrchestration := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"location": "Test Location",
			"temperature": map[string]float64{
				"celsius":    25.0,
				"fahrenheit": 77.0,
				"kelvin":     298.15,
			},
		})
	}))
	defer mockOrchestration.Close()

	handler := NewGatewayHandler(mockOrchestration.URL)

	// Test with valid CEP
	reqBody := CEPRequest{CEP: "29902555"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/cep", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ProcessCEP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check if response contains expected data
	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	if response["location"] != "Test Location" {
		t.Errorf("unexpected response: got %v", response)
	}
}

func TestGatewayHandler_ProcessCEP_InvalidJSON(t *testing.T) {
	handler := NewGatewayHandler("http://localhost:8080")

	req := httptest.NewRequest("POST", "/cep", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ProcessCEP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	var response ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	expectedMessage := "invalid request body"
	if response.Message != expectedMessage {
		t.Errorf("unexpected error message: got %v want %v", response.Message, expectedMessage)
	}
}

func TestGatewayHandler_HealthCheck(t *testing.T) {
	handler := NewGatewayHandler("http://localhost:8080")

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	handler.HealthCheck(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("unexpected status: got %v want %v", response["status"], "healthy")
	}

	if response["service"] != "otel-gateway" {
		t.Errorf("unexpected service name: got %v want %v", response["service"], "otel-gateway")
	}
}
