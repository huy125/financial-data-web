package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	model "github.com/huy125/financial-data-web/api/store/models"
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

type BalanceSheetMetadata struct {
	Symbol        string `json:"symbol"`
	AnnualReports []AnnualReport
}

type AnnualReport struct {
	FiscalDateEnding       string `json:"fiscalDateEnding"`
	TotalLiabilities       string `json:"totalLiabilities"`
	TotalShareholderEquity string `json:"totalShareholderEquity"`
}

type StockHandler struct {
	store  Store
	apiKey string
}

var metricMappings = map[string]string{
	"P/E Ratio":      "PERatio",
	"EPS":            "EPS",
	"Dividend Yield": "DividendYield",
	"MarketCap":      "MarketCapitalization",
	"RenewGrowth":    "QuarterlyRevenueGrowthYOY",
}

// GetStockBySymbolHandler fetches stock data for the given symbol.
func (h *StockHandler) GetStockBySymbolHandler(w http.ResponseWriter, r *http.Request) {
	if !r.URL.Query().Has("symbol") {
		http.Error(w, "Query parameter 'symbol' is required", http.StatusBadRequest)
		return
	}

	symbol := r.URL.Query().Get("symbol")
	symbol = strings.Trim(symbol, "\"")
	symbol = strings.Trim(symbol, "'")

	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout*time.Second)
	defer cancel()

	data, err := h.fetchStockData(ctx, symbol)
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

// GetOverviewStockBySymbolHandler fetches general company overview for the given symbol.
func (h *StockHandler) GetOverviewStockBySymbolHandler(w http.ResponseWriter, r *http.Request) {
	if !r.URL.Query().Has("symbol") {
		http.Error(w, "Query parameter 'symbol' is required", http.StatusBadRequest)
		return
	}

	symbol := r.URL.Query().Get("symbol")
	symbol = strings.Trim(symbol, "\"")
	symbol = strings.Trim(symbol, "'")

	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout*time.Second)
	defer cancel()

	data, err := h.fetchStockOverview(ctx, symbol)
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

func (h *StockHandler) fetchStockData(ctx context.Context, symbol string) (*TimeSeriesDaily, error) {
	return fetchDataFromAlphaVantage[TimeSeriesDaily](ctx, "TIME_SERIES_DAILY", symbol, h.apiKey)
}

func (h *StockHandler) fetchStockOverview(ctx context.Context, symbol string) (*map[string]string, error) {
	return fetchDataFromAlphaVantage[map[string]string](ctx, "OVERVIEW", symbol, h.apiKey)
}

func (h *StockHandler) fetchBalanceSheet(ctx context.Context, symbol string) (*BalanceSheetMetadata, error) {
	return fetchDataFromAlphaVantage[BalanceSheetMetadata](ctx, "BALANCE_SHEET", symbol, h.apiKey)
}

func fetchDataFromAlphaVantage[T any](ctx context.Context, function, symbol, apiKey string) (*T, error) {
	if apiKey == "" {
		log.Fatalf("ALPHA_VANTAGE_API_KEY environment variable is not set")
	}

	baseURL := &url.URL{
		Scheme: "https",
		Host:   "www.alphavantage.co",
		Path:   "/query",
	}

	q := baseURL.Query()
	q.Set("function", function)
	q.Set("symbol", symbol)
	q.Set("apikey", apiKey)
	baseURL.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error while constructing request for external api: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error while sending request to external api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error with status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error while reading response body: %w", err)
	}

	var data T
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	return &data, nil
}

func calculateDebtEquityRatio(_ context.Context, balanceSheet *BalanceSheetMetadata) (float64, error) {
	if len(balanceSheet.AnnualReports) == 0 {
		return 0, fmt.Errorf("no available data")
	}

	recentReport := balanceSheet.AnnualReports[0]

	totalLiabilities, err := strconv.ParseFloat(recentReport.TotalLiabilities, 64)
	if err != nil {
		return 0, err
	}

	totalEquity, err := strconv.ParseFloat(recentReport.TotalShareholderEquity, 64)
	if err != nil {
		return 0, err
	}

	if totalEquity == 0 {
		return 0, fmt.Errorf("cannot division by zero")
	}

	ratio := totalLiabilities / totalEquity
	return ratio, nil
}

func (h *StockHandler) updateStockMetrics(ctx context.Context, symbol string) ([]model.StockMetric, error) {
	var updatedStockMetrics []model.StockMetric
	saveStockMetric := func(metric model.Metric, value float64) error {
		now := time.Now()
		stockMetric := &model.StockMetric{
			ID:         uuid.New(),
			MetricID:   metric.ID,
			Value:      value,
			RecordedAt: now,
		}

		stockMetric, err := h.store.CreateStockMetric(ctx, stockMetric)
		if err != nil {
			return fmt.Errorf("failed to save metric %s for stock %s: %w", metric.Name, symbol, err)
		}

		updatedStockMetrics = append(updatedStockMetrics, *stockMetric)
		return nil
	}
	metrics, err := h.store.ListMetrics(ctx, 50, 0)
	if err != nil {
		return nil, err
	}

	overviewCh := make(chan *map[string]string, 1)
	balanceSheetCh := make(chan *BalanceSheetMetadata, 1)
	var overview *map[string]string
	var balanceSheet *BalanceSheetMetadata
	var overviewErr, balanceSheetErr error

	go func() {
		overview, overviewErr = h.fetchStockOverview(ctx, symbol)
		overviewCh <- overview
	}()
	go func() {
		balanceSheet, balanceSheetErr = h.fetchBalanceSheet(ctx, symbol)
		balanceSheetCh <- balanceSheet
	}()

	overviewData := <-overviewCh
	balanceSheetData := <-balanceSheetCh

	if overviewErr != nil && balanceSheetErr != nil {
		return nil, fmt.Errorf("failed to fetch stock metric information : %w", fmt.Errorf("overview error: %w, balance sheet error: %w", overviewErr, balanceSheetErr))
	}

	if balanceSheetErr == nil {
		for _, metric := range metrics {
			if metric.Name == "Debt/Equity Ratio" {
				// Calculate Debt/Equity ratio
				value, err := calculateDebtEquityRatio(ctx, balanceSheetData)
				if err != nil {
					log.Printf("failed to calculate Debt/Equity Ratio for stock %s: %v", symbol, err)
					continue
				}

				if err := saveStockMetric(metric, value); err != nil {
					log.Printf("falied to save stock metric: %v", err)
					continue
				}
			}
		}
	}

	if overviewErr == nil {
		overviewData := *overviewData
		for _, metric := range metrics {
			if metric.Name != "Debt/Equity Ratio" {
				// Process other metrics
				if externalFieldName, ok := metricMappings[metric.Name]; ok {
					if stringValue, ok := overviewData[externalFieldName]; ok {
						value, err := strconv.ParseFloat(stringValue, 64)
						if err != nil {
							log.Printf("Failed to parse value for metric %s from field %s: %v", metric.Name, externalFieldName, err)
							continue
						}

						if err := saveStockMetric(metric, value); err != nil {
							log.Printf("falied to save stock metric: %v", err)
							continue
						}
					}
				}
			}
		}
	}

	return updatedStockMetrics, nil
}
