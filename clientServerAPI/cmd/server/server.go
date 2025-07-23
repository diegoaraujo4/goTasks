package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

type ExchangeResponse struct {
	Rates struct {
		BRL float64 `json:"BRL"`
	} `json:"rates"`
	Base string `json:"base"`
	Date string `json:"date"`
}

// Legacy response structure for fallback APIs
type AwesomeAPIResponse struct {
	USDBRL struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

type Quote struct {
	Bid string `json:"bid"`
}

func initDB() (*sql.DB, error) {
	// Use different paths for Docker vs local development
	dbPath := "./quotes.db" // Default for local development
	if _, err := os.Stat("/data"); err == nil {
		// /data directory exists, we're in Docker
		dbPath = "/data/quotes.db"
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS quotes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		bid TEXT NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createTable)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func saveQuoteToDatabase(db *sql.DB, bid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	query := "INSERT INTO quotes (bid) VALUES (?)"

	done := make(chan error, 1)
	go func() {
		_, err := db.Exec(query, bid)
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			log.Printf("Error saving to database: %v", err)
			return err
		}
		return nil
	case <-ctx.Done():
		log.Printf("Database operation timeout: %v", ctx.Err())
		return ctx.Err()
	}
}

// Fallback function to try AwesomeAPI if ExchangeRate-API fails
func fetchFromAwesomeAPI() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("AwesomeAPI failed with status: %d", resp.StatusCode)
	}

	var apiResp AwesomeAPIResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResp)
	if err != nil {
		return "", err
	}

	return apiResp.USDBRL.Bid, nil
}

func fetchExchangeRateWithFallback() (*ExchangeResponse, error) {
	// Try primary API first (ExchangeRate-API)
	result, err := fetchExchangeRate()
	if err == nil {
		return result, nil
	}

	log.Printf("Primary ExchangeRate-API failed, trying AwesomeAPI fallback: %v", err)

	// Try AwesomeAPI as fallback
	bid, err := fetchFromAwesomeAPI()
	if err == nil {
		// Convert string bid to float64 then back to match our structure
		var brlRate float64
		if _, parseErr := fmt.Sscanf(bid, "%f", &brlRate); parseErr == nil {
			return &ExchangeResponse{
				Rates: struct {
					BRL float64 `json:"BRL"`
				}{
					BRL: brlRate,
				},
				Base: "USD",
				Date: time.Now().Format("2006-01-02"),
			}, nil
		}
	}

	return nil, fmt.Errorf("all exchange rate APIs failed - Primary: %v", err)
}

func fetchExchangeRate() (*ExchangeResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.exchangerate-api.com/v4/latest/USD", nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("API request timeout or error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		log.Printf("API returned non-200 status: %d %s", resp.StatusCode, resp.Status)
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	log.Printf("ExchangeRate-API Response Status: %s", resp.Status)

	var exchangeResp ExchangeResponse
	err = json.NewDecoder(resp.Body).Decode(&exchangeResp)
	if err != nil {
		log.Printf("Error decoding JSON response: %v", err)
		return nil, err
	}

	log.Printf("Successfully fetched BRL rate: %.4f", exchangeResp.Rates.BRL)
	return &exchangeResp, nil
}

func quotationHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		exchangeData, err := fetchExchangeRateWithFallback()
		if err != nil {
			log.Printf("Error fetching exchange rate from all sources: %v", err)
			http.Error(w, "Failed to fetch exchange rate", http.StatusInternalServerError)
			return
		}

		// Convert float64 to string with 4 decimal places
		bid := fmt.Sprintf("%.4f", exchangeData.Rates.BRL)
		log.Printf("Successfully fetched USD-BRL bid: %s", bid)

		// Save to database (with timeout handling)
		err = saveQuoteToDatabase(db, bid)
		if err != nil {
			log.Printf("Error saving quote to database: %v", err)
			// Continue serving the response even if DB save fails
		} else {
			log.Printf("Successfully saved quote to database: %s", bid)
		}

		quote := Quote{Bid: bid}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(quote)
		log.Printf("Response sent to client with bid: %s", bid)
	}
}

func main() {
	db, err := initDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	http.HandleFunc("/cotacao", quotationHandler(db))

	log.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
