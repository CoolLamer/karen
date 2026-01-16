-- Billing and usage tracking for monetization
-- Phase 1: Usage tracking foundation

-- Monthly usage tracking per tenant
CREATE TABLE tenant_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    period_start DATE NOT NULL,              -- First day of billing period
    period_end DATE NOT NULL,                -- Last day of billing period
    calls_count INT DEFAULT 0,               -- Number of calls in period
    minutes_used INT DEFAULT 0,              -- Total call duration in minutes
    time_saved_seconds INT DEFAULT 0,        -- Time saved (call duration + overhead)
    spam_calls_blocked INT DEFAULT 0,        -- Number of spam/marketing calls
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(tenant_id, period_start)
);

CREATE INDEX idx_tenant_usage_tenant ON tenant_usage(tenant_id);
CREATE INDEX idx_tenant_usage_period ON tenant_usage(period_start);

-- Per-call cost tracking for unit economics monitoring
CREATE TABLE call_costs (
    call_id UUID PRIMARY KEY REFERENCES calls(id) ON DELETE CASCADE,
    twilio_cost_cents INT DEFAULT 0,         -- Twilio voice charges
    stt_cost_cents INT DEFAULT 0,            -- Deepgram STT charges
    llm_cost_cents INT DEFAULT 0,            -- OpenAI LLM charges
    tts_cost_cents INT DEFAULT 0,            -- ElevenLabs TTS charges
    total_cost_cents INT DEFAULT 0,          -- Sum of all costs
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Add billing fields to tenants table
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS stripe_customer_id TEXT;
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS stripe_subscription_id TEXT;
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS trial_ends_at TIMESTAMPTZ;
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS current_period_start DATE;
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS current_period_calls INT DEFAULT 0;

-- Set trial_ends_at for existing trial tenants (14 days from creation)
UPDATE tenants
SET trial_ends_at = created_at + INTERVAL '14 days'
WHERE plan = 'trial' AND trial_ends_at IS NULL;

-- Trigger for tenant_usage updated_at
CREATE TRIGGER update_tenant_usage_updated_at BEFORE UPDATE ON tenant_usage
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
