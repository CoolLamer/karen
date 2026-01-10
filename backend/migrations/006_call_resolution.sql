-- Add call resolution tracking columns
-- These columns track when a call was first viewed and when it was marked as resolved

ALTER TABLE calls ADD COLUMN first_viewed_at timestamptz NULL;
ALTER TABLE calls ADD COLUMN resolved_at timestamptz NULL;
ALTER TABLE calls ADD COLUMN resolved_by uuid REFERENCES users(id) ON DELETE SET NULL;

-- Index for efficiently counting unresolved calls per tenant
CREATE INDEX idx_calls_resolution ON calls(tenant_id, resolved_at) WHERE resolved_at IS NULL;
