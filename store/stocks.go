package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Stock represents the stock schema in database.
type Stock struct {
	Model

	Symbol  string
	Company string
}

// StockMetric represents the join table between stock and metric schema in database.
type StockMetric struct {
	Model

	StockID    uuid.UUID
	MetricID   uuid.UUID
	Value      float64
}

// LatestStockMetric represents the most recent stock metric schema.
type LatestStockMetric struct {
	MetricName string
	Value      float64
	CreatedAt time.Time
	UpdatedAt time.Time
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
		RETURNING id, created_at, updated_at
	`

	err := s.db.pool.QueryRow(ctx, sql,
		stockMetric.StockID,
		stockMetric.MetricID,
		stockMetric.Value,
	).Scan(&stockMetric.ID, &stockMetric.CreatedAt, &stockMetric.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &stockMetric, nil
}

func (s *stockService) FindLastestStockMetrics(ctx context.Context, stockID uuid.UUID) ([]LatestStockMetric, error) {
	sql := `
			SELECT
				DISTINCT ON (sm.metric_id) 
				m.name AS metric_name,
				sm.value,
				sm.created_at,
				sm.updated_at
			FROM stock_metric sm
			INNER JOIN metric m ON sm.metric_id = m.id
			WHERE sm.stock_id = $1
			ORDER BY sm.metric_id, sm.created_at DESC
		`

	rows, err := s.db.pool.Query(ctx, sql, stockID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stockMetrics []LatestStockMetric
	for rows.Next() {
		var stockMetric LatestStockMetric
		if err := rows.Scan(&stockMetric.MetricName, &stockMetric.Value, &stockMetric.CreatedAt, &stockMetric.UpdatedAt); err != nil {
			return nil, err
		}
		stockMetrics = append(stockMetrics, stockMetric)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return stockMetrics, nil
}
