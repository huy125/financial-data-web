# Scoring System 🏆
## Description
This document represents how the scoring system works.
This system based on the group of several core metrics. Defining thresholds for each stock metric depends on the financial interpretation of that metric. Typically, thresholds represent ranges that categorize the metric into different levels of desirability or risk. Each metric also has a dedicated weighting which describes the importance of each metrics. It also makes the scoring logic flexible and extandable.
To be sum up, we can imagine that the scoring system is a function with following format:
$$
f(x) = \sum_{i=1}^{n} c_i x_i
$$
where:
- $i$ is the index representing each metric (e.g., P/E ratio, EPS, P/B ratio, etc.).
- $c_i$ is the weight assigned to each metric.
- $x_i$ is the value of the corresponding metric.

## 1. P/E Ratio (Price-to-Earnings)
The P/E ratio is a common measure of how expensive a stock is relative to its earnings. The general guidelines for P/E thresholds can be:
- **Low P/E (< 10)**: A stock is undervalued or has high growth potential. Usually more favorable for value investing.
- **Moderate P/E (10 - 20)**: The stock is priced fairly in relation to its earnings, neither undervalued nor overvalued.
- **High P/E (> 20)**: The stock is overvalued or expects high future growth, which can be risky.

## 2. EPS (Earnings Per Share)
EPS reflects the company’s profitability. Higher EPS usually indicates better profitability, but its value varies across industries.
- **High EPS (> 5)**: Strong profitability.
- **Moderate EPS (2 - 5)**: Decent profitability.
- **Low EPS (< 2)**: Lower profitability.

## 3. Debt-to-Equity (D/E) Ratio
The D/E ratio measures a company’s financial leverage. A higher ratio means the company relies more on debt.
- **Low D/E (< 0.5)**: A safe, conservative approach to using debt.
- **Moderate D/E (0.5 - 1.0)**: The company is using debt moderately.
- **High D/E (> 1.0)**: Riskier because the company is more heavily reliant on debt.

## 4. Dividend Yield
Dividend yield is important for income-focused investors. It measures how much a company pays out in dividends relative to its stock price.
- **High Yield (> 5%)**: A good yield, often from established companies.
- **Moderate Yield (3 - 5%)**: A balanced yield.
- **Low Yield (< 3%)**: Lower returns for dividend-seeking investors.

## 5. Market Capitalization
Market Cap represents the total market value of a company’s outstanding shares, reflecting its size and stability.  

- **Market Cap ≥ $100B** → **Score: 10** (Large Cap – Strong stability)  
- **$20B ≤ Market Cap < $100B** → **Score: 7** (Mid Cap – Moderate risk & growth potential)  
- **$2B ≤ Market Cap < $20B** → **Score: 5** (Small Cap – Higher risk)  
- **Market Cap < $2B** → **Score: 3** (Micro Cap – Very high risk)

## 6. Revenue Growth  
Revenue Growth measures a company's ability to increase its revenue over time, indicating business expansion.  

- **Revenue Growth ≥ 10%** → **Score: 10** (Strong growth)  
- **0% ≤ Revenue Growth < 10%** → **Score: 7** (Moderate growth)  
- **Revenue Growth < 0%** → **Score: 3** (Declining revenue) 

---
## **Conclusion**  

This scoring system helps evaluate stocks based on key financial indicators, with weighted contributions ensuring a balanced assessment.
The system remains flexible and adaptable for future enhancements.  
