package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Analysis represents the analysis schema in database.
type Analysis struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	StockID   uuid.UUID
	Score     float64
	CreatedAt time.Time
}

type analysisService struct {
	db *DB
}

func (s *analysisService) CreateAnalysis(ctx context.Context, analysis Analysis) (*Analysis, error) {
	sql := `
		INSERT INTO analysis (user_id, stock_id, score)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	err := s.db.pool.QueryRow(ctx, sql,
		analysis.UserID,
		analysis.StockID,
		analysis.Score,
	).Scan(&analysis.ID, &analysis.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &analysis, nil
}
