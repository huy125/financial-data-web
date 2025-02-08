package model

import (
	"time"

	"github.com/google/uuid"
)

// StockMetric represents the join table between stock and metric schema in database.
type StockMetric struct {
	ID         uuid.UUID
	StockID    uuid.UUID
	MetricID   uuid.UUID
	Value      float64
	RecordedAt time.Time
}
