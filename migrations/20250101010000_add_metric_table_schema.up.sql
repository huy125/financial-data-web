CREATE TABLE metric (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT
);

-- Insert basic metrics
INSERT INTO metric (name, description) VALUES
('P/E Ratio', 'Price-to-Earnings Ratio: A measure of valuation'),
('EPS', 'Earnings Per Share: A measure of profitability'),
('Market Cap', 'Market Capitalization: Total value of a company’s shares'),
('Revenue Growth', 'Growth in company revenue over time'),
('Dividend Yield', 'The dividend income relative to the stock price'),
('Debt/Equity Ratio', 'A measure of a company’s financial leverage');
