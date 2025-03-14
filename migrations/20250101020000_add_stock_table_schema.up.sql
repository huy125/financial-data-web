CREATE TABLE stock (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol VARCHAR(50) NOT NULL UNIQUE,
    company VARCHAR(255) NOT NULL
);

-- Insert basic stocks
INSERT INTO stock (symbol, company) VALUES
('AAPL', 'Apple Inc.'),
('MSFT', 'Microsoft Corporation'),
('GOOGL', 'Alphabet Inc.'),
('AMZN', 'Amazon.com Inc.'),
('TSLA', 'Tesla Inc.'),
('META', 'Meta Platforms Inc.'),
('NVDA', 'NVIDIA Corporation'),
('BRK.A', 'Berkshire Hathaway Inc.'),
('V', 'Visa Inc.'),
('JNJ', 'Johnson & Johnson'),
('XOM', 'Exxon Mobil Corporation'),
('PG', 'Procter & Gamble Co.'),
('JPM', 'JPMorgan Chase & Co.'),
('UNH', 'UnitedHealth Group Incorporated'),
('HD', 'The Home Depot Inc.'),
('MA', 'Mastercard Incorporated'),
('CVX', 'Chevron Corporation'),
('ABBV', 'AbbVie Inc.'),
('LLY', 'Eli Lilly and Company'),
('PEP', 'PepsiCo Inc.'),
('KO', 'The Coca-Cola Company'),
('MRK', 'Merck & Co., Inc.'),
('BAC', 'Bank of America Corporation'),
('PFE', 'Pfizer Inc.'),
('COST', 'Costco Wholesale Corporation'),
('AVGO', 'Broadcom Inc.'),
('TMO', 'Thermo Fisher Scientific Inc.'),
('DIS', 'The Walt Disney Company'),
('CSCO', 'Cisco Systems, Inc.'),
('NKE', 'Nike, Inc.'),
('ORCL', 'Oracle Corporation'),
('ABT', 'Abbott Laboratories'),
('INTC', 'Intel Corporation'),
('WMT', 'Walmart Inc.'),
('QCOM', 'QUALCOMM Incorporated'),
('ADBE', 'Adobe Inc.'),
('TM', 'Toyota Motor Corporation'),
('CMCSA', 'Comcast Corporation'),
('AMD', 'Advanced Micro Devices, Inc.'),
('CRM', 'Salesforce, Inc.'),
('HON', 'Honeywell International Inc.'),
('UPS', 'United Parcel Service, Inc.'),
('RTX', 'Raytheon Technologies Corporation'),
('MDT', 'Medtronic plc'),
('LIN', 'Linde plc'),
('MCD', 'McDonald''s Corporation'),
('TXN', 'Texas Instruments Incorporated'),
('NEE', 'NextEra Energy, Inc.'),
('PM', 'Philip Morris International Inc.'),
('UNP', 'Union Pacific Corporation');
