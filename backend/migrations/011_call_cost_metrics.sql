-- Add detailed metrics to call_costs table for cost calculation
-- These allow recalculating costs if pricing changes

-- Add raw metrics columns to call_costs
ALTER TABLE call_costs ADD COLUMN IF NOT EXISTS call_duration_seconds INT DEFAULT 0;
ALTER TABLE call_costs ADD COLUMN IF NOT EXISTS stt_duration_seconds INT DEFAULT 0;
ALTER TABLE call_costs ADD COLUMN IF NOT EXISTS llm_input_tokens INT DEFAULT 0;
ALTER TABLE call_costs ADD COLUMN IF NOT EXISTS llm_output_tokens INT DEFAULT 0;
ALTER TABLE call_costs ADD COLUMN IF NOT EXISTS tts_characters INT DEFAULT 0;

-- Add index on call_costs for tenant lookups via calls table
CREATE INDEX IF NOT EXISTS idx_call_costs_created ON call_costs(created_at);

-- Create a view for tenant cost summaries (joins calls to get tenant_id)
CREATE OR REPLACE VIEW tenant_cost_summary AS
SELECT
    c.tenant_id,
    DATE_TRUNC('month', cc.created_at) AS period,
    COUNT(*) AS call_count,
    SUM(cc.call_duration_seconds) AS total_duration_seconds,
    SUM(cc.twilio_cost_cents) AS twilio_cost_cents,
    SUM(cc.stt_cost_cents) AS stt_cost_cents,
    SUM(cc.llm_cost_cents) AS llm_cost_cents,
    SUM(cc.tts_cost_cents) AS tts_cost_cents,
    SUM(cc.total_cost_cents) AS total_api_cost_cents,
    SUM(cc.stt_duration_seconds) AS total_stt_seconds,
    SUM(cc.llm_input_tokens) AS total_llm_input_tokens,
    SUM(cc.llm_output_tokens) AS total_llm_output_tokens,
    SUM(cc.tts_characters) AS total_tts_characters
FROM call_costs cc
JOIN calls c ON c.id = cc.call_id
WHERE c.tenant_id IS NOT NULL
GROUP BY c.tenant_id, DATE_TRUNC('month', cc.created_at);
