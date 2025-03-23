DROP TRIGGER IF EXISTS update_analysis_updated_at ON analysis;

ALTER TABLE analysis
    DROP COLUMN IF EXISTS updated_at;