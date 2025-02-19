package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Stock struct {
	ID      uuid.UUID
	Symbol  string
	Company string
}

// StockMetric represents the join table between stock and metric schema in database.
type StockMetric struct {
	ID         uuid.UUID
	StockID    uuid.UUID
	MetricID   uuid.UUID
	Value      float64
	RecordedAt time.Time
}

type stockService struct {
	db *DB
}

func (s *stockService) Find(ctx context.Context, symbol string) (*Stock, error) {
	sql := "SELECT id, symbol, company FROM stock WHERE symbol = $1"
	var stock Stock

	err := s.db.pool.QueryRow(ctx, sql, symbol).Scan(
		&stock.ID,
		&stock.Symbol,
		&stock.Company,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return &stock, nil
}

func (s *stockService) CreateStockMetric(ctx context.Context, stockMetric StockMetric) (*StockMetric, error) {
	sql := `
		INSERT INTO stock_metric (stock_id, metric_id, value)
		VALUES ($1, $2, $3)
		RETURNING id, recorded_at
	`

	err := s.db.pool.QueryRow(ctx, sql,
		stockMetric.StockID,
		stockMetric.MetricID,
		stockMetric.Value,
	).Scan(&stockMetric.ID, &stockMetric.RecordedAt)
	if err != nil {
		return nil, err
	}

	return &stockMetric, nil
}
