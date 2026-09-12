CREATE TABLE IF NOT EXISTS links (
    code       TEXT PRIMARY KEY,
    target_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
