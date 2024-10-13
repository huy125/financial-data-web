package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// StockMetadata represents the metadata of a stock.
type StockMetadata struct {
	Information   string `json:"1. Information"`
	Symbol        string `json:"2. Symbol"`
	LastRefreshed string `json:"3. Last Refreshed"`
	OutputSize    string `json:"4. Output Size"`
	TimeZone      string `json:"5. Time Zone"`
}

// TimeSeriesDaily represents the daily time series of a stock.
type TimeSeriesDaily struct {
	StockMetadata `json:"Meta Data"`
	TimeSeries    map[string]map[string]string `json:"Time Series (Daily)"`
}

// GetStockBySymbolHandler fetches stock data for the given symbol.
func (s *Server) GetStockBySymbolHandler(w http.ResponseWriter, r *http.Request) {
	if !r.URL.Query().Has("symbol") {
		http.Error(w, "Query parameter 'symbol' is required", http.StatusBadRequest)
		return
	}

	symbol := r.URL.Query().Get("symbol")
	symbol = strings.Trim(symbol, "\"")
	symbol = strings.Trim(symbol, "'")

	data, err := s.fetchStockData(symbol)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, fmt.Sprintf("Stock data not found: %v", err), http.StatusNotFound)
			return
		}

		http.Error(w, fmt.Sprintf("Could not fetch stock data: %v", err), http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w,
			fmt.Sprintf("could not encode JSON: %v", err),
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonData)
	if err != nil {
		http.Error(w,
			fmt.Sprintf("error while writing response: %v", err),
			http.StatusInternalServerError,
		)
	}
}

func (s *Server) fetchStockData(symbol string) (*TimeSeriesDaily, error) {
	if s.apiKey == "" {
		log.Fatalf("ALPHA_VANTAGE_API_KEY environment variable is not set")
	}

	url := fmt.Sprintf("https://www.alphavantage.co/query?function=TIME_SERIES_DAILY&symbol=%s&apikey=%s", symbol, s.apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error while calling external api: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error with status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error while reading response body: %v", err)
	}

	var data TimeSeriesDaily
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	return &data, nil
}
