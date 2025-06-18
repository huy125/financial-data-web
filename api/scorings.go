package api

import (
	"encoding/json"
	"os"
)

// ScoringRange represents the lowest and highest values for a range, as well as its computed score.
type ScoringRange struct {
	Min   *float64 `json:"min"`
	Max   *float64 `json:"max"`
	Score float64  `json:"score"`
}

// Rule represents the threshold range and weight of each metric.
// The threshold ranges of a metric is a set of predefined value intervals used to evaluate the metric performance by
// assigning it a score.
type Rule struct {
	ScoringRanges []ScoringRange `json:"ranges"`
	Weight        float64        `json:"weight"`
}

// Config represents the scoring rules config.
type Config struct {
	Rules map[string]Rule `json:"rules"`
}

// applyFactors calculates the weighted score for a rule.
// If the value is outside the allowed range, it returns zero.
func applyFactors(value float64, rule Rule) float64 {
	for _, r := range rule.ScoringRanges {
		if r.Min != nil && value < *r.Min {
			continue
		}

		if r.Max != nil && value > *r.Max {
			continue
		}

		return float64(r.Score) * rule.Weight
	}
	return 0
}

func loadScoringRules(filename string) (map[string]Rule, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, err
	}

	return config.Rules, nil
}
