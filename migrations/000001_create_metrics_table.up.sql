CREATE TABLE IF NOT EXISTS metrics (
    id      TEXT PRIMARY KEY,
    m_type  TEXT NOT NULL CHECK (m_type IN ('gauge', 'counter')),
    delta   BIGINT,
    value   DOUBLE PRECISION,
    hash    TEXT
);