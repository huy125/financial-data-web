DROP TRIGGER IF EXISTS update_stock_metric_updated_at ON stock_metric;

ALTER TABLE stock_metric
    DROP COLUMN IF EXISTS updated_at;