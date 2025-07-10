package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Quote struct {
	Bid string `json:"bid"`
}

func fetchQuoteFromServer() (*Quote, error) {
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
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Server request timeout or error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var quote Quote
	err = json.Unmarshal(body, &quote)
	if err != nil {
		return nil, err
	}

	return &quote, nil
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
	quote, err := fetchQuoteFromServer()
	if err != nil {
		log.Fatal("Failed to fetch quote from server:", err)
	}

	err = saveQuoteToFile(quote.Bid)
	if err != nil {
		log.Fatal("Failed to save quote to file:", err)
	}

	fmt.Printf("Current USD/BRL exchange rate: %s\n", quote.Bid)
}
