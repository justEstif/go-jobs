CREATE EXTENSION IF NOT EXISTS pg_trgm;

ALTER TABLE companies ADD COLUMN normalized_name TEXT NOT NULL DEFAULT '';

CREATE INDEX idx_companies_normalized_trgm ON companies USING gin (normalized_name gin_trgm_ops);

-- Backfill existing companies: lowercase, strip legal suffixes, remove non-alphanumeric.
UPDATE companies SET normalized_name = lower(trim(
    regexp_replace(
        regexp_replace(
            name,
            '\s*(,?\s*)(Inc\.?|LLC|Ltd\.?|Corp\.?|Corporation|Co\.?|Company|Group|Holdings|PLC|GmbH|S\.A\.?|Pty\.?)\s*$',
            '',
            'gi'
        ),
        '[^a-z0-9 ]', '', 'gi'
    )
));
