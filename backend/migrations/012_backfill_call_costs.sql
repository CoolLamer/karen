-- Backfill call costs for existing calls that have duration but no cost records
-- This migration calculates estimated costs based on call duration for:
--   - Twilio: 0.85 cents/minute
--   - Deepgram STT: 0.77 cents/minute (using call duration as proxy)
-- LLM and TTS costs remain 0 as we don't have historical metrics for those.

-- Insert cost records for calls that don't have them
INSERT INTO call_costs (
    call_id,
    twilio_cost_cents,
    stt_cost_cents,
    llm_cost_cents,
    tts_cost_cents,
    total_cost_cents,
    call_duration_seconds,
    stt_duration_seconds,
    llm_input_tokens,
    llm_output_tokens,
    tts_characters
)
SELECT
    c.id AS call_id,
    -- Twilio: 0.85 cents/minute, rounded to nearest cent
    ROUND(EXTRACT(EPOCH FROM (c.ended_at - c.started_at)) / 60.0 * 0.85)::INT AS twilio_cost_cents,
    -- Deepgram STT: 0.77 cents/minute, using call duration as proxy
    ROUND(EXTRACT(EPOCH FROM (c.ended_at - c.started_at)) / 60.0 * 0.77)::INT AS stt_cost_cents,
    -- LLM: 0 (no historical data)
    0 AS llm_cost_cents,
    -- TTS: 0 (no historical data)
    0 AS tts_cost_cents,
    -- Total: Twilio + STT
    ROUND(EXTRACT(EPOCH FROM (c.ended_at - c.started_at)) / 60.0 * 0.85)::INT +
    ROUND(EXTRACT(EPOCH FROM (c.ended_at - c.started_at)) / 60.0 * 0.77)::INT AS total_cost_cents,
    -- Duration in seconds
    EXTRACT(EPOCH FROM (c.ended_at - c.started_at))::INT AS call_duration_seconds,
    EXTRACT(EPOCH FROM (c.ended_at - c.started_at))::INT AS stt_duration_seconds,
    -- Metrics: 0 (no historical data)
    0 AS llm_input_tokens,
    0 AS llm_output_tokens,
    0 AS tts_characters
FROM calls c
WHERE c.ended_at IS NOT NULL
  AND c.started_at IS NOT NULL
  AND c.ended_at > c.started_at
  AND NOT EXISTS (
      SELECT 1 FROM call_costs cc WHERE cc.call_id = c.id
  );

-- Update existing call_costs records that have zero costs but the call has duration
-- This handles cases where a record exists but was never populated
UPDATE call_costs cc
SET
    twilio_cost_cents = ROUND(EXTRACT(EPOCH FROM (c.ended_at - c.started_at)) / 60.0 * 0.85)::INT,
    stt_cost_cents = ROUND(EXTRACT(EPOCH FROM (c.ended_at - c.started_at)) / 60.0 * 0.77)::INT,
    total_cost_cents = ROUND(EXTRACT(EPOCH FROM (c.ended_at - c.started_at)) / 60.0 * 0.85)::INT +
                       ROUND(EXTRACT(EPOCH FROM (c.ended_at - c.started_at)) / 60.0 * 0.77)::INT,
    call_duration_seconds = EXTRACT(EPOCH FROM (c.ended_at - c.started_at))::INT,
    stt_duration_seconds = EXTRACT(EPOCH FROM (c.ended_at - c.started_at))::INT
FROM calls c
WHERE cc.call_id = c.id
  AND cc.total_cost_cents = 0
  AND c.ended_at IS NOT NULL
  AND c.started_at IS NOT NULL
  AND c.ended_at > c.started_at;
