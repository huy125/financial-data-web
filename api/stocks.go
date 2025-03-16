package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	lctx "github.com/hamba/logger/v2/ctx"
	"github.com/huy125/financial-data-web/api/mapper"
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

// ThresholdRange represents the lowest, highest value and score of each range of threshold.
type ThresholdRange struct {
	Min   float64
	Max   float64
	Score float64
}

// ScoringRules reprepsents threshold range and weigh of each metrics.
type ScoringRule struct {
	Ranges []ThresholdRange
	Weight float64
}

type fetchResult struct {
	overview                     *OverviewMetadata
	balanceSheet                 *BalanceSheetMetadata
	overviewErr, balanceSheetErr error
}

const (
	highScore    = 10
	mediumScore  = 7
	lowScore     = 5
	veryLowScore = 3
)

const (
	StrongBuy = 8
	Buy       = 6
	Hold      = 4
	Sell      = 2
)

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

	stock, err := s.store.FindStockBySymbol(ctx, symbol)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			http.Error(w, "Stock data is not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	stockMetrics, err := s.updateStockMetrics(ctx, stock)
	if err != nil {
		http.Error(w,
			"Internal server error",
			http.StatusInternalServerError,
		)
		return
	}

	score, err := s.scoreStock(ctx, stock)
	if err != nil {
		s.log.Error("Failed to scoring stock", lctx.Error("error", err))
		http.Error(w,
			"Internal server error",
			http.StatusInternalServerError,
		)
		return
	}

	action := getAction(score)
	confidenceLevel := calculateConfidenceLevel(stockMetrics)

	// A temporary userID while waiting authentification system
	userID := uuid.MustParse("b6ce2ebd-e6ae-4618-9d15-0ecb9f04a72e")
	analysis, err := s.store.CreateAnalysis(ctx, userID, stock.ID, score)
	if err != nil {
		s.log.Error("Failed to create analysis", lctx.Error("error", err))
		http.Error(w,
			"Internal server error",
			http.StatusInternalServerError,
		)
		return
	}

	reason := ""
	recommendation, err := s.store.CreateRecommendation(ctx, analysis.ID, action, confidenceLevel, reason)
	if err != nil {
		s.log.Error("Failed to create recommendation", lctx.Error("error", err))
		http.Error(w,
			"Internal server error",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(mapper.ToAPIRecommendation(recommendation))
	if err != nil {
		http.Error(w, "Failed to encode the response", http.StatusInternalServerError)
		return
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

func (s *Server) updateStockMetrics(ctx context.Context, stock *store.Stock) ([]store.StockMetric, error) {
	const maxNumMetric = 50
	const offset = 0
	metrics, err := s.store.ListMetrics(ctx, maxNumMetric, offset)
	if err != nil {
		return nil, err
	}

	metricMap := buildMetricMap(metrics)
	overview, balanceSheet, fetchErr := s.combineStockData(ctx, stock.Symbol)
	if fetchErr != nil {
		return nil, fetchErr
	}

	var updatedStockMetrics []store.StockMetric
	saveStockMetric := s.saveStockMetric(ctx, stock, &updatedStockMetrics)

	if balanceSheet != nil {
		s.processBalanceSheetMetrics(ctx, balanceSheet, metricMap, saveStockMetric)
	}

	if overview != nil {
		s.processOverviewMetrics(ctx, overview, metricMap, saveStockMetric)
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

func (s *Server) processBalanceSheetMetrics(
	_ context.Context,
	balanceSheet *BalanceSheetMetadata,
	metricMap map[string]store.Metric,
	save func(store.Metric, float64),
) {
	if metric, ok := metricMap["Debt/Equity Ratio"]; ok {
		value, err := calculateDebtEquityRatio(balanceSheet)
		if err != nil {
			s.log.Error("failed to calculate Debt/Equity Ratio for", lctx.Error("error", err))
			return
		}
		save(metric, value)
	}
}

func (s *Server) processOverviewMetrics(
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
				s.log.Error("failed to convert metric to float", lctx.Error("error", err))
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
			s.log.Error("failed to save metric", lctx.Str("metric", metric.Name), lctx.Error("error", err))
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

func (s *Server) scoreStock(ctx context.Context, stock *store.Stock) (float64, error) {
	totalScore := 0.0

	stockMetrics, err := s.store.FindLastestStockMetrics(ctx, stock.ID)
	if err != nil {
		return 0.0, err
	}

	rules := getScoringRules()
	for _, stockMetric := range stockMetrics {
		if rule, exists := rules[stockMetric.MetricName]; exists {
			totalScore += calculateScore(stockMetric.Value, rule)
		}
	}

	return totalScore, nil
}

func calculateScore(value float64, rule ScoringRule) float64 {
	for _, r := range rule.Ranges {
		if value >= r.Min && value < r.Max {
			return float64(r.Score) * rule.Weight
		}
	}
	return 0
}

func getScoringRules() map[string]ScoringRule {
	return map[string]ScoringRule{
		"P/E Ratio":         getPERatioRule(),
		"EPS":               getEPSRule(),
		"Revenue Growth":    getRevenueGrowthRule(),
		"Debt/Equity Ratio": getDebtEquityRule(),
		"Dividend Yield":    getDividendYieldRule(),
		"Market Cap":        getMarketCapRule(),
	}
}

func getPERatioRule() ScoringRule {
	const (
		peLow    = 10
		peMedium = 20
		peHigh   = 30
		peWeight = 0.16
	)
	return newScoringRule(peWeight, []ThresholdRange{
		{Min: 0, Max: peLow, Score: highScore},
		{Min: peLow, Max: peMedium, Score: mediumScore},
		{Min: peMedium, Max: peHigh, Score: lowScore},
		{Min: peHigh, Max: math.Inf(1), Score: veryLowScore},
	})
}

func getEPSRule() ScoringRule {
	const (
		epsLow    = 2
		epsHigh   = 5
		epsWeight = 0.12
	)
	return newScoringRule(epsWeight, []ThresholdRange{
		{Min: epsHigh, Max: math.Inf(1), Score: highScore},
		{Min: epsLow, Max: epsHigh, Score: mediumScore},
		{Min: 0, Max: epsLow, Score: veryLowScore},
	})
}

func getRevenueGrowthRule() ScoringRule {
	const (
		revenueGrowthLow    = 0.1
		revenueGrowthHigh   = 0.2
		revenueGrowthWeight = 0.24
	)
	return newScoringRule(revenueGrowthWeight, []ThresholdRange{
		{Min: revenueGrowthHigh, Max: math.Inf(1), Score: highScore},
		{Min: revenueGrowthLow, Max: revenueGrowthHigh, Score: mediumScore},
		{Min: math.Inf(-1), Max: 0, Score: veryLowScore},
	})
}

func getDebtEquityRule() ScoringRule {
	const (
		debtEquityLow    = 0
		debtEquityMedium = 0.5
		debtEquityHigh   = 1.0
		debtEquityWeight = 0.16
	)
	return newScoringRule(debtEquityWeight, []ThresholdRange{
		{Min: debtEquityLow, Max: debtEquityMedium, Score: highScore},
		{Min: debtEquityMedium, Max: debtEquityHigh, Score: mediumScore},
		{Min: debtEquityHigh, Max: math.Inf(1), Score: veryLowScore},
	})
}

func getDividendYieldRule() ScoringRule {
	const (
		dividendLow    = 0
		dividendMedium = 0.03
		dividendHigh   = 0.05
		dividendWeight = 0.08
	)
	return newScoringRule(dividendWeight, []ThresholdRange{
		{Min: dividendHigh, Max: math.Inf(1), Score: highScore},
		{Min: dividendMedium, Max: dividendHigh, Score: mediumScore},
		{Min: dividendLow, Max: dividendMedium, Score: veryLowScore},
	})
}

func getMarketCapRule() ScoringRule {
	const (
		smallCap        = 2_000_000_000
		midCap          = 20_000_000_000
		largeCap        = 100_000_000_000
		marketCapWeight = 0.24
	)
	return newScoringRule(marketCapWeight, []ThresholdRange{
		{Min: largeCap, Max: math.Inf(1), Score: highScore},
		{Min: midCap, Max: largeCap, Score: mediumScore},
		{Min: smallCap, Max: midCap, Score: lowScore},
		{Min: 0, Max: smallCap, Score: veryLowScore},
	})
}

func newScoringRule(weight float64, ranges []ThresholdRange) ScoringRule {
	return ScoringRule{Weight: weight, Ranges: ranges}
}

func getAction(score float64) store.Action {
	switch {
	case score >= StrongBuy:
		return store.StrongBuy
	case score >= Buy:
		return store.Buy
	case score >= Hold:
		return store.Hold
	case score >= Sell:
		return store.Sell
	default:
		return store.StrongSell
	}
}

func calculateConfidenceLevel(stockMetrics []store.StockMetric) float64 {
	totalMetrics := len(stockMetrics)
	if totalMetrics == 0 {
		return 0
	}

	normalizedMetrics := minMaxNormalizeMetrics(stockMetrics)

	var sum, varianceSum float64
	for _, value := range normalizedMetrics {
		sum += value
	}
	mean := sum / float64(totalMetrics)

	for _, value := range normalizedMetrics {
		varianceSum += math.Pow(value-mean, 2)
	}
	sampleVariance := varianceSum / float64(totalMetrics-1)

	stdDeviation := math.Sqrt(sampleVariance)
	confidence := 100 - math.Min(stdDeviation*100, 100)

	return confidence
}

func minMaxNormalizeMetrics(stockMetrics []store.StockMetric) []float64 {
	minValue := stockMetrics[0].Value
	maxValue := stockMetrics[0].Value

	for _, metric := range stockMetrics {
		if metric.Value < minValue {
			minValue = metric.Value
		}
		if metric.Value > maxValue {
			maxValue = metric.Value
		}
	}

	// Normalize each metric using Min-Max formula
	normalizedValues := make([]float64, len(stockMetrics))
	for i, metric := range stockMetrics {
		normalizedValues[i] = (metric.Value - minValue) / (maxValue - minValue)
	}

	return normalizedValues
}
