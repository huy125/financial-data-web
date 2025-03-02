package store

import (
	"context"

	"github.com/google/uuid"
)

// Metric represents the metric model
type Metric struct {
	ID          uuid.UUID
	Name        string
	Description string
}

type metricService struct {
	db *DB
}

func (s *metricService) ListMetrics(ctx context.Context, limit, offset int) ([]Metric, error) {
	sql := "SELECT id, name, description FROM metric LIMIT $1 OFFSET $2"
	rows, err := s.db.pool.Query(ctx, sql, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []Metric
	for rows.Next() {
		var metric Metric
		if err := rows.Scan(&metric.ID, &metric.Name, &metric.Description); err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return metrics, nil
}
