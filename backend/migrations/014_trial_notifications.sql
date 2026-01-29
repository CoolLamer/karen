-- Migration 014: Trial Lifecycle Notifications
-- Adds tracking fields for trial conversion prompts, grace period notifications, and phone number release

-- Add trial notification tracking fields to tenants table
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS trial_day10_notification_sent_at TIMESTAMPTZ;
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS trial_day12_notification_sent_at TIMESTAMPTZ;
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS trial_day14_notification_sent_at TIMESTAMPTZ;
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS trial_grace_notification_sent_at TIMESTAMPTZ;
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS phone_number_released_at TIMESTAMPTZ;

-- Index for finding tenants needing trial lifecycle processing
-- This helps the background job efficiently find tenants at each stage
CREATE INDEX IF NOT EXISTS idx_tenants_trial_lifecycle
ON tenants(plan, trial_ends_at)
WHERE plan = 'trial' AND trial_ends_at IS NOT NULL;

-- Add trial lifecycle config values to global_config
INSERT INTO global_config (key, value, description) VALUES
    ('trial_grace_period_days', '7', 'Days after trial expiration before releasing phone number'),
    ('sms_sender_number', '', 'Twilio phone number for sending SMS notifications (E.164 format, e.g., +420123456789)')
ON CONFLICT (key) DO NOTHING;
