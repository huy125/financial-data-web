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
	"sync"
	"time"

	"github.com/huy125/financial-data-web/api/store"
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

// GetStockAnalysisBySymbolHandler fetches general company overview for the given symbol.
func (h *StockHandler) GetStockAnalysisBySymbolHandler(w http.ResponseWriter, r *http.Request) {
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

	_, err = h.updateStockMetrics(ctx, symbol)
	if err != nil {
		http.Error(w,
			fmt.Sprintf("could not update stock metric: %v", err),
			http.StatusInternalServerError,
		)
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
		return 0, errors.New("no available data")
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
		return 0, errors.New("cannot division by zero")
	}

	ratio := totalLiabilities / totalEquity
	return ratio, nil
}

func (h *StockHandler) updateStockMetrics(ctx context.Context, symbol string) ([]model.StockMetric, error) {
	stock, err := h.store.FindStockBySymbol(ctx, symbol)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, errors.New("stock not found")
		}
		return nil, fmt.Errorf("error while fetching stock by symbol: %w", err)
	}

	const maxNumMetric = 50
	const offset = 0
	metrics, err := h.store.ListMetrics(ctx, maxNumMetric, offset)
	if err != nil {
		return nil, err
	}

	metricMap := buildMetricMap(metrics)
	overview, balanceSheet, fetchErr := h.combineStockData(ctx, symbol)
	if fetchErr != nil {
		return nil, fetchErr
	}

	var updatedStockMetrics []model.StockMetric
	addMetric := h.metricAdder(ctx, stock, &updatedStockMetrics)

	if balanceSheet != nil {
		processBalanceSheetMetrics(ctx, balanceSheet, metricMap, addMetric, symbol)
	}

	if overview != nil {
		processOverviewMetrics(ctx, *overview, metricMap, addMetric)
	}

	return updatedStockMetrics, nil
}

func (h *StockHandler) combineStockData(
	ctx context.Context,
	symbol string,
) (*map[string]string, *BalanceSheetMetadata, error) {
	type fetchResult struct {
		overview     *map[string]string
		balanceSheet *BalanceSheetMetadata
		err          error
	}

	const numCh = 2
	resultCh := make(chan fetchResult, numCh)
	var wg sync.WaitGroup
	wg.Add(numCh)

	go func() {
		defer wg.Done()
		overview, err := h.fetchStockOverview(ctx, symbol)
		resultCh <- fetchResult{overview: overview, err: err}
	}()

	go func() {
		defer wg.Done()
		balanceSheet, err := h.fetchBalanceSheet(ctx, symbol)
		resultCh <- fetchResult{balanceSheet: balanceSheet, err: err}
	}()

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	var overview *map[string]string
	var balanceSheet *BalanceSheetMetadata
	var fetchErrors []error

	for res := range resultCh {
		if res.err != nil {
			fetchErrors = append(fetchErrors, res.err)
		}
		if res.overview != nil {
			overview = res.overview
		}

		if res.balanceSheet != nil {
			balanceSheet = res.balanceSheet
		}
	}

	if len(fetchErrors) == numCh {
		return nil, nil, fmt.Errorf("failed to fetch stock data: %w & %w", fetchErrors[0], fetchErrors[1])
	}

	return overview, balanceSheet, nil
}

func processBalanceSheetMetrics(
	ctx context.Context,
	balanceSheet *BalanceSheetMetadata,
	metricMap map[string]model.Metric,
	save func(model.Metric, float64),
	symbol string,
) {
	if metric, ok := metricMap["Debt/Equity Ratio"]; ok {
		value, err := calculateDebtEquityRatio(ctx, balanceSheet)
		if err != nil {
			log.Printf("failed to calculate Debt/Equity Ratio for stock %s: %v", symbol, err)
			return
		}
		save(metric, value)
	}
}

func processOverviewMetrics(
	_ context.Context,
	overview map[string]string,
	metricMap map[string]model.Metric,
	save func(model.Metric, float64),
) {
	metricMapping := getMetricMapping()

	for name, metric := range metricMap {
		if name == "Debt/Equity Ratio" {
			continue
		}
		if field, ok := metricMapping[name]; ok {
			if valueStr, ok := overview[field]; ok {
				value, err := strconv.ParseFloat(valueStr, 64)
				if err != nil {
					log.Printf("failed to parse value for metric %s: %v", name, err)
					continue
				}
				save(metric, value)
			}
		}
	}
}

func (h *StockHandler) metricAdder(
	ctx context.Context,
	stock *model.Stock,
	updatedMetrics *[]model.StockMetric,
) func(metric model.Metric, value float64) {
	return func(metric model.Metric, value float64) {
		stockMetric := &model.StockMetric{
			StockID:    stock.ID,
			MetricID:   metric.ID,
			Value:      value,
			RecordedAt: time.Now(),
		}
		savedMetric, err := h.store.CreateStockMetric(ctx, stockMetric)
		if err != nil {
			log.Printf("failed to save metric %s for stock %s: %v", metric.Name, stock.Symbol, err)
			return
		}
		*updatedMetrics = append(*updatedMetrics, *savedMetric)
	}
}

func buildMetricMap(metrics []model.Metric) map[string]model.Metric {
	metricMap := make(map[string]model.Metric)
	for _, metric := range metrics {
		metricMap[metric.Name] = metric
	}
	return metricMap
}

func getMetricMapping() map[string]string {
	return map[string]string{
		"P/E Ratio":      "PERatio",
		"EPS":            "EPS",
		"Dividend Yield": "DividendYield",
		"Market Cap":     "MarketCapitalization",
		"Revenue Growth": "QuarterlyRevenueGrowthYOY",
	}
}
