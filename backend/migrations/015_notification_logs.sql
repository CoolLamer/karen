-- Migration 015: Notification audit logs
-- Track all SMS and push notifications sent to users for admin auditing

CREATE TABLE IF NOT EXISTS notification_logs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    channel text NOT NULL,              -- 'sms' or 'apns'
    notification_type text NOT NULL,    -- e.g. 'trial_day10', 'call_completed', 'grace_warning'
    recipient text NOT NULL,            -- phone number (SMS) or device token prefix (APNs)
    tenant_id uuid REFERENCES tenants(id) ON DELETE SET NULL,
    body text,                          -- message content
    status text NOT NULL DEFAULT 'sent', -- 'sent' or 'failed'
    error_message text,                 -- error details if failed
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_notification_logs_created_at ON notification_logs(created_at DESC);
CREATE INDEX idx_notification_logs_tenant_id ON notification_logs(tenant_id, created_at DESC);
CREATE INDEX idx_notification_logs_channel ON notification_logs(channel);
