package api

import (
	"encoding/json"
	"errors"
	"math"
	"os"
)

// Range represents its lowest, highest value and assinging score of each range.
type Range struct {
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Score float64 `json:"score"`
}

// Rule represents threshold range and weigh of each metrics.
type Rule struct {
	Ranges []Range `json:"ranges"`
	Weight float64 `json:"weight"`
}

// Config represents the scoring rules config.
type Config struct {
	Rules map[string]Rule `json:"rules"`
}

func applyFactors(value float64, rule Rule) float64 {
	for _, r := range rule.Ranges {
		if value < r.Min || value >= r.Max {
			continue
		}

		return float64(r.Score) * rule.Weight
	}
	return 0
}

func convertInfToMathInf(rules map[string]Rule) error {
	const infinityThreshold = 1e+308

	for _, rule := range rules {
		for i := range rule.Ranges {
			if rule.Ranges[i].Max == -infinityThreshold || rule.Ranges[i].Min == infinityThreshold {
				return errors.New("syntax error in config template")
			}
			if rule.Ranges[i].Max == infinityThreshold {
				rule.Ranges[i].Max = math.Inf(1)
			}

			if rule.Ranges[i].Min == -infinityThreshold {
				rule.Ranges[i].Min = math.Inf(-1)
			}
		}
	}

	return nil
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

	err = convertInfToMathInf(config.Rules)
	if err != nil {
		return nil, err
	}

	return config.Rules, nil
}
