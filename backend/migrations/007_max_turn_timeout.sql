-- Add configurable max_turn_timeout_ms to tenants
-- This is the hard timeout (in ms) for waiting for speech_final from STT
-- Default 4000ms matches the current hardcoded constant

ALTER TABLE tenants ADD COLUMN max_turn_timeout_ms INTEGER DEFAULT NULL;
-- NULL means use system default (4000ms)
