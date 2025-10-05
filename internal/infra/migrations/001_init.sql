CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 2) Create enum type only if missing (Postgres has no IF NOT EXISTS for enums)
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'message_status') THEN
    CREATE TYPE message_status AS ENUM ('queued', 'processing', 'sent', 'failed');
  END IF;
END$$;

-- 3) Ensure all enum labels exist (safe no-ops if present)
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_enum e JOIN pg_type t ON t.oid = e.enumtypid
    WHERE t.typname = 'message_status' AND e.enumlabel = 'queued'
  ) THEN
    ALTER TYPE message_status ADD VALUE 'queued';
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_enum e JOIN pg_type t ON t.oid = e.enumtypid
    WHERE t.typname = 'message_status' AND e.enumlabel = 'processing'
  ) THEN
    ALTER TYPE message_status ADD VALUE 'processing';
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_enum e JOIN pg_type t ON t.oid = e.enumtypid
    WHERE t.typname = 'message_status' AND e.enumlabel = 'sent'
  ) THEN
    ALTER TYPE message_status ADD VALUE 'sent';
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_enum e JOIN pg_type t ON t.oid = e.enumtypid
    WHERE t.typname = 'message_status' AND e.enumlabel = 'failed'
  ) THEN
    ALTER TYPE message_status ADD VALUE 'failed';
  END IF;
END$$;

-- 4) Table
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

-- 5) Indexes
CREATE INDEX IF NOT EXISTS idx_messages_status_created ON messages (status, created_at);
CREATE INDEX IF NOT EXISTS idx_messages_sent_at ON messages (sent_at);