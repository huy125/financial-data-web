package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type StockMetaData struct {
	Information 	string `json:"1. Information"`
	Symbol 		string `json:"2. Symbol"`
	LastRefreshed 	string `json:"3. Last Refreshed"`
	OutputSize	string `json:"4. Output Size"`
	TimeZone	string `json:"5. Time Zone"`
}

type TimeSeriesDaily struct {
	StockMetaData 	`json:"Meta Data"`
	TimeSeries 	map[string]map[string]string `json:"Time Series (Daily)"`
}

const symbol string = "IBM"

func GetStockHandler(w http.ResponseWriter, r *http.Request) {
	data, err := fetchStockData(symbol)

	if err != nil {
        	http.Error(w, fmt.Sprintf("error fetching stock data: %v", err), http.StatusBadRequest)
    	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, fmt.Sprintf("error while encoding JSON: %v", err), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func fetchStockData(symbol string) (*TimeSeriesDaily, error) {
	apiKey := os.Getenv("ALPHA_VANTAGE_API_KEY")
	if apiKey == "" {
		log.Fatalf("ALPHA_VANTAGE_API_KEY environment variable is not set")
	}

	url := fmt.Sprintf("https://www.alphavantage.co/query?function=TIME_SERIES_DAILY&symbol=%s&apikey=%s", symbol, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error while calling external api: %v", err)
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
