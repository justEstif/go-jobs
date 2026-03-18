-- HTTP session store for alexedwards/scs pgxstore.
-- Separate from the CLI sessions table (token + user_id) which is used by
-- the opaque token pattern for CLI auth. This table stores encrypted session
-- blobs for the web layer.
CREATE TABLE IF NOT EXISTS http_sessions (
    token  TEXT PRIMARY KEY,
    data   BYTEA NOT NULL,
    expiry TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_http_sessions_expiry ON http_sessions(expiry);
