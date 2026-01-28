-- Add rejection_reason column to store specific reason when calls are rejected due to limits
-- Possible values: "trial_expired", "limit_exceeded", "subscription_cancelled", "subscription_suspended"
ALTER TABLE calls ADD COLUMN IF NOT EXISTS rejection_reason TEXT;
