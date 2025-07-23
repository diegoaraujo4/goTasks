package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Quote struct {
	Bid string `json:"bid"`
}

const maxRetries = 5

func fetchQuoteFromServer() (*Quote, error) {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("Attempt %d/%d to fetch quote from server", attempt, maxRetries)

		quote, err := fetchQuoteFromServerOnce()
		if err == nil && isValidQuote(quote) {
			log.Printf("Successfully fetched quote on attempt %d: %s", attempt, quote.Bid)
			return quote, nil
		}

		lastErr = err
		if err != nil {
			log.Printf("Attempt %d failed: %v", attempt, err)
		} else {
			log.Printf("Attempt %d failed: invalid or empty bid received", attempt)
		}

		// Wait before retrying (exponential backoff)
		if attempt < maxRetries {
			waitTime := time.Duration(attempt) * 2 * time.Second
			log.Printf("Waiting %v before retry...", waitTime)
			time.Sleep(waitTime)
		}
	}

	return nil, fmt.Errorf("failed to fetch valid quote after %d attempts, last error: %v", maxRetries, lastErr)
}

func fetchQuoteFromServerOnce() (*Quote, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	// Use different hostnames for Docker vs local development
	serverURL := "http://localhost:8080/cotacao" // Default for local development
	if _, err := os.Stat("/data"); err == nil {
		// /data directory exists, we're in Docker
		serverURL = "http://server:8080/cotacao"
	}

	req, err := http.NewRequestWithContext(ctx, "GET", serverURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("server request timeout or error: %v", err)
	}
	defer resp.Body.Close()

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s, body: %s", resp.StatusCode, resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Check if response body is empty
	if len(body) == 0 {
		return nil, fmt.Errorf("server returned empty response")
	}

	var quote Quote
	err = json.Unmarshal(body, &quote)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %v, body: %s", err, string(body))
	}

	return &quote, nil
}

func isValidQuote(quote *Quote) bool {
	if quote == nil {
		return false
	}

	// Check if bid is empty or contains only whitespace
	if strings.TrimSpace(quote.Bid) == "" {
		return false
	}

	// Additional validation: check if bid looks like a number
	// This is a basic check - you might want more sophisticated validation
	if quote.Bid == "0" || quote.Bid == "0.0" || quote.Bid == "0.00" {
		return false
	}

	return true
}

func saveQuoteToFile(bid string) error {
	content := fmt.Sprintf("Dólar: %s", bid)

	// Use different paths for Docker vs local development
	filePath := "./cotacao.txt" // Default for local development
	if _, err := os.Stat("/data"); err == nil {
		// /data directory exists, we're in Docker
		filePath = "/data/cotacao.txt"
	}

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		log.Printf("Error saving quote to file: %v", err)
		return err
	}

	log.Printf("Quote saved to cotacao.txt: %s", content)
	return nil
}

func main() {
	log.Println("Starting client to fetch USD/BRL exchange rate...")

	quote, err := fetchQuoteFromServer()
	if err != nil {
		log.Printf("Failed to fetch quote from server after %d attempts: %v", maxRetries, err)
		log.Fatal("Exiting due to repeated failures")
		os.Exit(1)
	}

	log.Printf("Successfully obtained exchange rate: %s", quote.Bid)

	err = saveQuoteToFile(quote.Bid)
	if err != nil {
		log.Fatal("Failed to save quote to file:", err)
	}

	fmt.Printf("Current USD/BRL exchange rate: %s\n", quote.Bid)
	log.Println("Client completed successfully")
}
