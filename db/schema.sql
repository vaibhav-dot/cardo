-- Cardo schema. Loaded automatically when the Postgres container first starts.
CREATE EXTENSION IF NOT EXISTS vector;

-- Raw rows exactly as the bank exported them.
CREATE TABLE IF NOT EXISTS raw_transactions (
    id          BIGSERIAL PRIMARY KEY,
    date        DATE           NOT NULL,
    description TEXT           NOT NULL,
    amount      NUMERIC(12, 2) NOT NULL
);

-- One row per cleaned-up merchant string. The categoriser fills this in,
-- so each unique description is only sent to the LLM once (cheap + fast).
CREATE TABLE IF NOT EXISTS merchant_cache (
    description TEXT PRIMARY KEY,
    merchant    TEXT,
    category    TEXT,
    source      TEXT CHECK (source IN ('rule', 'llm', 'manual')),
    updated_at  TIMESTAMPTZ DEFAULT now()
);

-- What the agent queries: raw rows joined to their category.
CREATE OR REPLACE VIEW transactions AS
SELECT r.id, r.date, r.description, r.amount, m.merchant, m.category
FROM raw_transactions r
LEFT JOIN merchant_cache m USING (description);

-- Day 8–9: the deals catalogue for RAG.
CREATE TABLE IF NOT EXISTS offers (
    id        BIGSERIAL PRIMARY KEY,
    title     TEXT NOT NULL,
    body      TEXT NOT NULL,
    category  TEXT,
    merchant  TEXT,
    url       TEXT,
    embedding vector(1024)  -- match your embedding model's dimension
);

-- Day 10: timed suggestions shown in the insight feed.
CREATE TABLE IF NOT EXISTS nudges (
    id         BIGSERIAL PRIMARY KEY,
    show_on    DATE NOT NULL,
    message    TEXT NOT NULL,
    reason     TEXT NOT NULL,
    offer_id   BIGINT REFERENCES offers(id),
    created_at TIMESTAMPTZ DEFAULT now()
);
