package mapper

import (
	"strconv"

	"github.com/huy125/financial-data-web/api/dto"
	"github.com/huy125/financial-data-web/store"
)

// ToAPIUser converts a store analysis to an API analysis.
func ToAPIRecommendation(r *store.Recommendation) *dto.RecommendationDto {
	return &dto.RecommendationDto{
		ID:              r.ID.String(),
		AnalysisID:      r.AnalysisID.String(),
		Action:          string(r.Action),
		ConfidenceLevel: strconv.FormatFloat(r.ConfidenceLevel, 'f', 2, 64),
		Reason:          r.Reason,
	}
}
