CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE message_status AS ENUM ('queued', 'processing', 'sent', 'failed');

CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    to_phone VARCHAR(32) NOT NULL,
    content TEXT NOT NULL CHECK (char_length(content) <= 1000),
    status message_status NOT NULL DEFAULT 'queued',
    retry_count INT NOT NULL DEFAULT 0,
    provider_message_id VARCHAR(128),
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sent_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_messages_status_created ON messages (status, created_at);
CREATE INDEX IF NOT EXISTS idx_messages_sent_at ON messages (sent_at);
