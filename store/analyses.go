package store

import (
	"context"

	"github.com/google/uuid"
)

// Analysis represents the analysis schema in database.
type Analysis struct {
	Model

	UserID  uuid.UUID
	StockID uuid.UUID
	Score   float64
}

type analysisService struct {
	db *DB
}

func (s *analysisService) Create(ctx context.Context, analysis *Analysis) (*Analysis, error) {
	sql := `
		INSERT INTO analysis (user_id, stock_id, score)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	err := s.db.pool.QueryRow(ctx, sql,
		analysis.UserID,
		analysis.StockID,
		analysis.Score,
	).Scan(&analysis.ID, &analysis.CreatedAt, &analysis.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return analysis, nil
}
