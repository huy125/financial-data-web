package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Action string
const (
	StrongBuy Action = "strong_buy"
	Buy Action = "buy"
	Hold Action = "hold"
	Sell Action = "sell"
	StrongSell Action = "strong_sell"
)

type Recommendation struct {
	ID uuid.UUID
	AnalysisID uuid.UUID
	Action Action
	ConfidenceLevel float64
	Reason string
	CreatedAt time.Time
}

type recommendationService struct {
	db *DB
}
// Save recommendation
func (r *recommendationService) Create(ctx context.Context, recommendation *Recommendation) (*Recommendation, error) {
	sql := `
		INSERT INTO recommendation (analysis_id, action, confidence_score, reason)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	err := r.db.pool.QueryRow(ctx, sql,
		recommendation.AnalysisID,
		recommendation.Action,
		recommendation.ConfidenceLevel,
		recommendation.Reason,
	).Scan(&recommendation.ID, &recommendation.CreatedAt)
	if err != nil {
		return nil, err
	}

	return recommendation, nil

}