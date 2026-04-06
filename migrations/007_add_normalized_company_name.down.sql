DROP INDEX IF EXISTS idx_companies_normalized_trgm;
ALTER TABLE companies DROP COLUMN IF EXISTS normalized_name;
DROP EXTENSION IF EXISTS pg_trgm;
