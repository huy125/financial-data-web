package dto

// User represents user information.
type AnalysisDto struct {
	ID      string `json:"id"`
	StockID string `json:"stock_id"`
	UserID  string `json:"user_id"`
	Score   string `json:"score"`
}
