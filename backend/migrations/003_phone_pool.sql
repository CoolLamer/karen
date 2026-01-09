-- Migration: Phone number pool
-- This allows phone numbers to be pre-provisioned and auto-assigned during onboarding

-- Ensure tenant_id is nullable (it already is, but let's be explicit)
-- Numbers with NULL tenant_id are available for assignment

-- Add an index for finding available numbers efficiently
CREATE INDEX IF NOT EXISTS idx_tenant_phone_numbers_available
ON tenant_phone_numbers(tenant_id) WHERE tenant_id IS NULL;

-- To add phone numbers to the pool, insert them with NULL tenant_id:
-- INSERT INTO tenant_phone_numbers (twilio_number, twilio_sid)
-- VALUES ('+420123456789', 'PN_twilio_sid_here');
