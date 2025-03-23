package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	lctx "github.com/hamba/logger/v2/ctx"
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

type recommendationResp struct {
	ID              string `json:"id"`
	AnalysisID      string `json:"analysis_id"`
	Action          string `json:"action"`
	ConfidenceLevel float64 `json:"confidence_level"`
	Reason          string `json:"reason"`
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

// GetStockAnalysisBySymbolHandler performs a basic stock evaluation.
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

	recommendation, err := s.analyzeStock(ctx, stock)
	if err != nil {
		s.log.Error("Failed to create recommendation", lctx.Error("error", err))
		http.Error(w,
			"Internal server error",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(toRecommendationResp(recommendation))
	if err != nil {
		http.Error(w, "Failed to encode the response", http.StatusInternalServerError)
		return
	}
}

func (s *Server) analyzeStock(ctx context.Context, stock *store.Stock) (*store.Recommendation, error) {
	stockMetrics, err := s.updateStockMetrics(ctx, stock)
	if err != nil {
		return nil, err
	}

	score, err := s.scoreStock(ctx, stock)
	if err != nil {
		return nil, err
	}

	action := getAction(score)
	confidenceLevel := calculateConfidenceLevel(stockMetrics)

	userID := uuid.MustParse("b6ce2ebd-e6ae-4618-9d15-0ecb9f04a72e") // Temporary user ID
	analysis, err := s.store.CreateAnalysis(ctx, userID, stock.ID, score)
	if err != nil {
		return nil, err
	}

	recommendation, err := s.store.CreateRecommendation(ctx, analysis.ID, action, confidenceLevel, "")
	if err != nil {
		return nil, err
	}

	return recommendation, nil
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
	var totalScore float64

	stockMetrics, err := s.store.FindLastestStockMetrics(ctx, stock.ID)
	if err != nil {
		return totalScore, err
	}

	rules := getScoringRules()
	for _, stockMetric := range stockMetrics {
		rule, exists := rules[stockMetric.MetricName]
		if !exists {
			continue
		}
		
		totalScore += applyFactors(stockMetric.Value, rule)
	}

	return totalScore, nil
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

	const exponent = 2
	const maxPercentage = 100
	normalizedMetrics := minMaxNormalizeMetrics(stockMetrics)

	var sum, varianceSum float64
	for _, value := range normalizedMetrics {
		sum += value
	}
	mean := sum / float64(totalMetrics)

	for _, value := range normalizedMetrics {
		varianceSum += math.Pow(value-mean, exponent)
	}
	sampleVariance := varianceSum / float64(totalMetrics-1)

	stdDeviation := math.Sqrt(sampleVariance)
	confidence := maxPercentage - math.Min(stdDeviation*maxPercentage, maxPercentage)

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

func toRecommendationResp(r *store.Recommendation) recommendationResp {
	return recommendationResp{
		ID:              r.ID.String(),
		AnalysisID:      r.AnalysisID.String(),
		Action:          string(r.Action),
		ConfidenceLevel: r.ConfidenceLevel,
		Reason:          r.Reason,
	}
}
