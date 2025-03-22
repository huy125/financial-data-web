DROP TRIGGER IF EXISTS update_metric_updated_at ON metric;

ALTER TABLE metric
    DROP COLUMN IF EXISTS created_at,
    DROP COLUMN IF EXISTS updated_at;