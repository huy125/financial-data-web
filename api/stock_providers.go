package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

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

	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, ErrNotFound
	case http.StatusOK:
		// continue
	default:
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
