-- Migration 013: Global Config and Robocall Detection
-- Adds global_config table for admin-editable settings and robocall tracking to calls

-- Global config table for admin-editable system settings
CREATE TABLE IF NOT EXISTS global_config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    description TEXT,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Insert default global config values
INSERT INTO global_config (key, value, description) VALUES
    -- Base turn timeout (used when adaptive is disabled, or as starting point for adaptive)
    ('max_turn_timeout_ms', '4000', 'Maximum time to wait for speech_final before forcing finalization (ms)'),

    -- Adaptive turn settings
    ('adaptive_turn_enabled', 'true', 'Enable adaptive turn timeout based on sentence completion'),
    ('adaptive_min_timeout_ms', '500', 'Minimum timeout when sentence ends with punctuation (ms)'),
    ('adaptive_text_decay_rate_ms', '15', 'Timeout reduction per character of buffered text (ms)'),
    ('adaptive_sentence_end_bonus_ms', '1500', 'Additional reduction when text ends with .!? (ms)'),

    -- Robocall detection settings
    ('robocall_detection_enabled', 'true', 'Enable automatic robocall detection'),
    ('robocall_max_call_duration_ms', '300000', 'Maximum call duration before auto-hangup (5 minutes)'),
    ('robocall_silence_threshold_ms', '30000', 'Silence threshold to detect robocalls (30 seconds)'),
    ('robocall_barge_in_threshold', '3', 'Number of rapid barge-ins to trigger robocall detection'),
    ('robocall_barge_in_window_ms', '15000', 'Time window for counting rapid barge-ins (15 seconds)'),
    ('robocall_repetition_threshold', '3', 'Number of identical phrases to trigger robocall detection'),
    ('robocall_hold_keywords', '["nezavěšujte","do not hang up","please hold","moment prosím","čekejte prosím"]', 'JSON array of hold music/robocall indicator keywords')
ON CONFLICT (key) DO NOTHING;

-- Add robocall tracking columns to calls table
ALTER TABLE calls ADD COLUMN IF NOT EXISTS is_robocall BOOLEAN DEFAULT FALSE;
ALTER TABLE calls ADD COLUMN IF NOT EXISTS robocall_reason TEXT;

-- Index for robocall analysis
CREATE INDEX IF NOT EXISTS idx_calls_is_robocall ON calls(is_robocall) WHERE is_robocall = TRUE;
