-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS calls (
  id uuid PRIMARY KEY,
  provider text NOT NULL,
  provider_call_id text NOT NULL,
  from_number text NOT NULL,
  to_number text NOT NULL,
  status text NOT NULL,
  started_at timestamptz NOT NULL,
  ended_at timestamptz NULL,
  UNIQUE (provider, provider_call_id)
);

CREATE TABLE IF NOT EXISTS call_utterances (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  call_id uuid NOT NULL REFERENCES calls(id) ON DELETE CASCADE,
  speaker text NOT NULL,
  text text NOT NULL,
  sequence int NOT NULL,
  started_at timestamptz NULL,
  ended_at timestamptz NULL,
  stt_confidence double precision NULL,
  interrupted boolean NOT NULL DEFAULT false,
  UNIQUE (call_id, sequence)
);

CREATE TABLE IF NOT EXISTS call_screening_results (
  call_id uuid PRIMARY KEY REFERENCES calls(id) ON DELETE CASCADE,
  legitimacy_label text NOT NULL,
  legitimacy_confidence double precision NOT NULL DEFAULT 0,
  intent_category text NOT NULL DEFAULT '',
  intent_text text NOT NULL DEFAULT '',
  entities_json jsonb NOT NULL DEFAULT '{}'::jsonb,
  needs_follow_up boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now()
);


