CREATE TABLE IF NOT EXISTS companies (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    careers_url TEXT NOT NULL DEFAULT '',
    ats_type    TEXT NOT NULL,
    scrape_type TEXT NOT NULL DEFAULT 'http',
    board_token TEXT NOT NULL,
    active      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (ats_type, board_token)
);

CREATE INDEX IF NOT EXISTS idx_companies_ats_token ON companies(ats_type, board_token);
CREATE INDEX IF NOT EXISTS idx_companies_active ON companies(active);
