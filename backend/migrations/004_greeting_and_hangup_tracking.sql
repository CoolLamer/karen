-- Add hangup tracking to calls table
ALTER TABLE calls ADD COLUMN ended_by TEXT NULL;

COMMENT ON COLUMN calls.ended_by IS 'Who initiated the call termination: agent, caller, or NULL for unknown/legacy calls';
