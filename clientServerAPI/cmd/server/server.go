package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

type ExchangeResponse struct {
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

func fetchExchangeRate() (*ExchangeResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
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

	var exchangeResp ExchangeResponse
	err = json.NewDecoder(resp.Body).Decode(&exchangeResp)
	if err != nil {
		return nil, err
	}

	return &exchangeResp, nil
}

func quotationHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		exchangeData, err := fetchExchangeRate()
		if err != nil {
			log.Printf("Error fetching exchange rate: %v", err)
			http.Error(w, "Failed to fetch exchange rate", http.StatusInternalServerError)
			return
		}

		bid := exchangeData.USDBRL.Bid

		// Save to database (with timeout handling)
		err = saveQuoteToDatabase(db, bid)
		if err != nil {
			log.Printf("Error saving quote to database: %v", err)
			// Continue serving the response even if DB save fails
		}

		quote := Quote{Bid: bid}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(quote)
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
