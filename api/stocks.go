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

	"github.com/huy125/financial-data-web/store"
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

// OverviewMetada represents the overall financial information of a stock.
type OverviewMetadata struct {
	Symbol                    string `json:"symbol"`
	MarketCapitalization      string `json:"MarketCapitalization"`
	PERatio                   string `json:"PERatio"`
	EPS                       string `json:"EPS"`
	DividendYield             string `json:"DividendYield"`
	QuarterlyRevenueGrowthYOY string `json:"QuarterlyRevenueGrowthYOY"`
}

// BalanceSheetMetadata represents a summary of the financial balances of a stock.
type BalanceSheetMetadata struct {
	Symbol        string `json:"symbol"`
	AnnualReports []AnnualReport
}

// AnnualReport represents key financial data from a stock's annual balance sheet report.
type AnnualReport struct {
	FiscalDateEnding       string `json:"fiscalDateEnding"`
	TotalLiabilities       string `json:"totalLiabilities"`
	TotalShareholderEquity string `json:"totalShareholderEquity"`
}

type fetchResult struct {
	overview                     *OverviewMetadata
	balanceSheet                 *BalanceSheetMetadata
	overviewErr, balanceSheetErr error
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

	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout*time.Second)
	defer cancel()

	data, err := s.fetchStockData(ctx, symbol)
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
func (s *Server) GetStockAnalysisBySymbolHandler(w http.ResponseWriter, r *http.Request) {
	if !r.URL.Query().Has("symbol") {
		http.Error(w, "Query parameter 'symbol' is required", http.StatusBadRequest)
		return
	}

	symbol := r.URL.Query().Get("symbol")
	symbol = strings.Trim(symbol, "\"")
	symbol = strings.Trim(symbol, "'")

	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout*time.Second)
	defer cancel()

	data, err := s.fetchStockOverview(ctx, symbol)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, fmt.Sprintf("Stock data not found: %v", err), http.StatusNotFound)
			return
		}

		http.Error(w, fmt.Sprintf("Could not fetch stock data: %v", err), http.StatusInternalServerError)
		return
	}

	_, err = s.updateStockMetrics(ctx, symbol)
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

func (s *Server) fetchStockData(ctx context.Context, symbol string) (*TimeSeriesDaily, error) {
	return fetchDataFromAlphaVantage[TimeSeriesDaily](ctx, "TIME_SERIES_DAILY", symbol, s.apiKey)
}

func (s *Server) fetchStockOverview(ctx context.Context, symbol string) (*OverviewMetadata, error) {
	return fetchDataFromAlphaVantage[OverviewMetadata](ctx, "OVERVIEW", symbol, s.apiKey)
}

func (s *Server) fetchBalanceSheet(ctx context.Context, symbol string) (*BalanceSheetMetadata, error) {
	return fetchDataFromAlphaVantage[BalanceSheetMetadata](ctx, "BALANCE_SHEET", symbol, s.apiKey)
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

func calculateDebtEquityRatio(balanceSheet *BalanceSheetMetadata) (float64, error) {
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
		return 0, errors.New("total equity should not be zero")
	}

	ratio := totalLiabilities / totalEquity
	return ratio, nil
}

func (s *Server) updateStockMetrics(ctx context.Context, symbol string) ([]store.StockMetric, error) {
	stock, err := s.store.FindStockBySymbol(ctx, symbol)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, errors.New("stock not found")
		}
		return nil, fmt.Errorf("error while fetching stock by symbol: %w", err)
	}

	const maxNumMetric = 50
	const offset = 0
	metrics, err := s.store.ListMetrics(ctx, maxNumMetric, offset)
	if err != nil {
		return nil, err
	}

	metricMap := buildMetricMap(metrics)
	overview, balanceSheet, fetchErr := s.combineStockData(ctx, symbol)
	if fetchErr != nil {
		return nil, fetchErr
	}

	var updatedStockMetrics []store.StockMetric
	saveStockMetric := s.saveStockMetric(ctx, stock, &updatedStockMetrics)

	if balanceSheet != nil {
		processBalanceSheetMetrics(ctx, balanceSheet, metricMap, saveStockMetric, symbol)
	}

	if overview != nil {
		processOverviewMetrics(ctx, overview, metricMap, saveStockMetric)
	}

	return updatedStockMetrics, nil
}

func (s *Server) combineStockData(
	ctx context.Context,
	symbol string,
) (*OverviewMetadata, *BalanceSheetMetadata, error) {
	const fetches = 2
	resultCh := make(chan fetchResult, fetches)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		overview, overviewErr := s.fetchStockOverview(ctx, symbol)
		resultCh <- fetchResult{overview: overview, overviewErr: overviewErr}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		balanceSheet, balanceSheetErr := s.fetchBalanceSheet(ctx, symbol)
		resultCh <- fetchResult{balanceSheet: balanceSheet, balanceSheetErr: balanceSheetErr}
	}()

	wg.Wait()
	close(resultCh)

	var overview *OverviewMetadata
	var balanceSheet *BalanceSheetMetadata
	var overviewErr, balanceSheetErr error

	for res := range resultCh {
		if res.overview != nil {
			overview = res.overview
		}

		if res.balanceSheet != nil {
			balanceSheet = res.balanceSheet
		}

		if res.overviewErr != nil {
			overviewErr = res.overviewErr
		}

		if res.balanceSheetErr != nil {
			balanceSheetErr = res.balanceSheetErr
		}
	}

	if overviewErr != nil && balanceSheetErr != nil {
		return nil, nil, errors.Join(overviewErr, balanceSheetErr)
	}

	return overview, balanceSheet, nil
}

func processBalanceSheetMetrics(
	_ context.Context,
	balanceSheet *BalanceSheetMetadata,
	metricMap map[string]store.Metric,
	save func(store.Metric, float64),
	symbol string,
) {
	if metric, ok := metricMap["Debt/Equity Ratio"]; ok {
		value, err := calculateDebtEquityRatio(balanceSheet)
		if err != nil {
			log.Printf("failed to calculate Debt/Equity Ratio for stock %s: %v", symbol, err)
			return
		}
		save(metric, value)
	}
}

func processOverviewMetrics(
	_ context.Context,
	overview *OverviewMetadata,
	metricMap map[string]store.Metric,
	save func(store.Metric, float64),
) {
	metricExtractors := map[string]func(*OverviewMetadata) string{
		"P/E Ratio":      func(o *OverviewMetadata) string { return o.PERatio },
		"EPS":            func(o *OverviewMetadata) string { return o.EPS },
		"Dividend Yield": func(o *OverviewMetadata) string { return o.DividendYield },
		"Market Cap":     func(o *OverviewMetadata) string { return o.MarketCapitalization },
		"Revenue Growth": func(o *OverviewMetadata) string { return o.QuarterlyRevenueGrowthYOY },
	}

	for name, metricModel := range metricMap {
		if extractor, exists := metricExtractors[name]; exists {
			value, err := strconv.ParseFloat(extractor(overview), 64)
			if err != nil {
				log.Printf("failed to convert metric %s to float: %v", name, err)
			}
			save(metricModel, value)
		}
	}
}

func (s *Server) saveStockMetric(
	ctx context.Context,
	stock *store.Stock,
	updatedMetrics *[]store.StockMetric,
) func(metric store.Metric, value float64) {
	return func(metric store.Metric, value float64) {
		savedMetric, err := s.store.CreateStockMetric(ctx, stock.ID, metric.ID, value)
		if err != nil {
			log.Printf("failed to save metric %s for stock %s: %v", metric.Name, stock.Symbol, err)
			return
		}
		*updatedMetrics = append(*updatedMetrics, *savedMetric)
	}
}

func buildMetricMap(metrics []store.Metric) map[string]store.Metric {
	metricMap := make(map[string]store.Metric, len(metrics))
	for _, metric := range metrics {
		metricMap[metric.Name] = metric
	}
	return metricMap
}
