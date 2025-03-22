package store

import (
	"context"

	"github.com/google/uuid"
)

type Action string

const (
	StrongBuy  Action = "Strong Buy"
	Buy        Action = "Buy"
	Hold       Action = "Hold"
	Sell       Action = "Sell"
	StrongSell Action = "Strong Sell"
)

type Recommendation struct {
	Model

	AnalysisID      uuid.UUID
	Action          Action
	ConfidenceLevel float64
	Reason          string
}

type recommendationService struct {
	db *DB
}

func (r *recommendationService) Create(ctx context.Context, recommendation *Recommendation) (*Recommendation, error) {
	sql := `
		INSERT INTO recommendation (analysis_id, action, confidence_level, reason)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	err := r.db.pool.QueryRow(ctx, sql,
		recommendation.AnalysisID,
		recommendation.Action,
		recommendation.ConfidenceLevel,
		recommendation.Reason,
	).Scan(&recommendation.ID, &recommendation.CreatedAt, &recommendation.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return recommendation, nil
}
