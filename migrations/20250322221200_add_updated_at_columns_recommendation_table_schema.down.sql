DROP TRIGGER IF EXISTS update_recommendation_updated_at ON recommendation;

ALTER TABLE recommendation
    DROP COLUMN IF EXISTS updated_at;
