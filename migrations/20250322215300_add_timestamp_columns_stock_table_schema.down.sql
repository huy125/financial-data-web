DROP TRIGGER IF EXISTS update_stock_updated_at ON stock;

ALTER TABLE stock
    DROP COLUMN IF EXISTS created_at,
    DROP COLUMN IF EXISTS updated_at;