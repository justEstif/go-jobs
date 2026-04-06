CREATE TABLE IF NOT EXISTS contacts (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                 UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    first_name              TEXT NOT NULL,
    last_name               TEXT NOT NULL,
    full_name               TEXT GENERATED ALWAYS AS (first_name || ' ' || last_name) STORED,
    email                   TEXT NOT NULL DEFAULT '',
    title                   TEXT NOT NULL DEFAULT '',
    linkedin_url            TEXT NOT NULL DEFAULT '',
    connected_on            DATE,
    company_name            TEXT NOT NULL,
    normalized_company_name TEXT NOT NULL DEFAULT '',
    company_id              UUID REFERENCES companies(id),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_contacts_user_linkedin
    ON contacts(user_id, linkedin_url) WHERE linkedin_url <> '';

CREATE INDEX idx_contacts_user_company
    ON contacts(user_id, company_id) WHERE company_id IS NOT NULL;

CREATE INDEX idx_contacts_normalized_company
    ON contacts(normalized_company_name);
