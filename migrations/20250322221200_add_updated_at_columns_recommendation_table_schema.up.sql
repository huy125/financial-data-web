ALTER TABLE recommendation
	ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;

CREATE TRIGGER update_recommendation_updated_at
    BEFORE UPDATE ON stock_metric
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();