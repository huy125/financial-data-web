package mapper

import (
	"strconv"

	"github.com/huy125/financial-data-web/api/dto"
	"github.com/huy125/financial-data-web/store"
)

// ToAPIUser converts a store analysis to an API analysis.
func ToAPIAnalysis(a *store.Analysis) *dto.AnalysisDto {
	return &dto.AnalysisDto{
		ID:      a.ID.String(),
		UserID:  a.UserID.String(),
		StockID: a.StockID.String(),
		Score:   strconv.FormatFloat(a.Score, 'f', 2, 64),
	}
}
