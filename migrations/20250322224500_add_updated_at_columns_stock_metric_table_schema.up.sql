ALTER TABLE stock_metric
	ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;

CREATE TRIGGER update_stock_metric_updated_at
    BEFORE UPDATE ON stock_metric
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();