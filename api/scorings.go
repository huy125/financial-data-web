package api

import "math"

// ThresholdRange represents the lowest, highest value and score of each range of threshold.
type ThresholdRange struct {
	Min   float64
	Max   float64
	Score float64
}

// ScoringRules reprepsents threshold range and weigh of each metrics.
type ScoringRule struct {
	Ranges []ThresholdRange
	Weight float64
}

func applyFactors(value float64, rule ScoringRule) float64 {
	for _, r := range rule.Ranges {
		if value < r.Min || value >= r.Max {
			continue
		}
		
		return float64(r.Score) * rule.Weight
	}
	return 0
}

func getScoringRules() map[string]ScoringRule {
	return map[string]ScoringRule{
		"P/E Ratio":         getPERatioRule(),
		"EPS":               getEPSRule(),
		"Revenue Growth":    getRevenueGrowthRule(),
		"Debt/Equity Ratio": getDebtEquityRule(),
		"Dividend Yield":    getDividendYieldRule(),
		"Market Cap":        getMarketCapRule(),
	}
}

func getPERatioRule() ScoringRule {
	const (
		peLow    = 10
		peMedium = 20
		peHigh   = 30
		peWeight = 0.16
	)
	return newScoringRule(peWeight, []ThresholdRange{
		{Min: 0, Max: peLow, Score: highScore},
		{Min: peLow, Max: peMedium, Score: mediumScore},
		{Min: peMedium, Max: peHigh, Score: lowScore},
		{Min: peHigh, Max: math.Inf(1), Score: veryLowScore},
	})
}

func getEPSRule() ScoringRule {
	const (
		epsLow    = 2
		epsHigh   = 5
		epsWeight = 0.12
	)
	return newScoringRule(epsWeight, []ThresholdRange{
		{Min: epsHigh, Max: math.Inf(1), Score: highScore},
		{Min: epsLow, Max: epsHigh, Score: mediumScore},
		{Min: 0, Max: epsLow, Score: veryLowScore},
	})
}

func getRevenueGrowthRule() ScoringRule {
	const (
		revenueGrowthLow    = 0.1
		revenueGrowthHigh   = 0.2
		revenueGrowthWeight = 0.24
	)
	return newScoringRule(revenueGrowthWeight, []ThresholdRange{
		{Min: revenueGrowthHigh, Max: math.Inf(1), Score: highScore},
		{Min: revenueGrowthLow, Max: revenueGrowthHigh, Score: mediumScore},
		{Min: math.Inf(-1), Max: 0, Score: veryLowScore},
	})
}

func getDebtEquityRule() ScoringRule {
	const (
		debtEquityLow    = 0
		debtEquityMedium = 0.5
		debtEquityHigh   = 1.0
		debtEquityWeight = 0.16
	)
	return newScoringRule(debtEquityWeight, []ThresholdRange{
		{Min: debtEquityLow, Max: debtEquityMedium, Score: highScore},
		{Min: debtEquityMedium, Max: debtEquityHigh, Score: mediumScore},
		{Min: debtEquityHigh, Max: math.Inf(1), Score: veryLowScore},
	})
}

func getDividendYieldRule() ScoringRule {
	const (
		dividendLow    = 0
		dividendMedium = 0.03
		dividendHigh   = 0.05
		dividendWeight = 0.08
	)
	return newScoringRule(dividendWeight, []ThresholdRange{
		{Min: dividendHigh, Max: math.Inf(1), Score: highScore},
		{Min: dividendMedium, Max: dividendHigh, Score: mediumScore},
		{Min: dividendLow, Max: dividendMedium, Score: veryLowScore},
	})
}

func getMarketCapRule() ScoringRule {
	const (
		smallCap        = 2_000_000_000
		midCap          = 20_000_000_000
		largeCap        = 100_000_000_000
		marketCapWeight = 0.24
	)
	return newScoringRule(marketCapWeight, []ThresholdRange{
		{Min: largeCap, Max: math.Inf(1), Score: highScore},
		{Min: midCap, Max: largeCap, Score: mediumScore},
		{Min: smallCap, Max: midCap, Score: lowScore},
		{Min: 0, Max: smallCap, Score: veryLowScore},
	})
}

func newScoringRule(weight float64, ranges []ThresholdRange) ScoringRule {
	return ScoringRule{Weight: weight, Ranges: ranges}
}
