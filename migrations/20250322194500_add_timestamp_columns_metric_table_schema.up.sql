ALTER TABLE metric
	ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
	ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;

CREATE TRIGGER update_metric_updated_at
    BEFORE UPDATE ON metric
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();