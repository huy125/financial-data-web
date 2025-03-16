package dto

// User represents user information.
type RecommendationDto struct {
	ID string `json:"id"`
	AnalysisID string `json:"analysis_id"`
	Action string `json:"action"`
	ConfidenceLevel string `json:"confidence_level"`
	Reason string `json:"reason"`
}
