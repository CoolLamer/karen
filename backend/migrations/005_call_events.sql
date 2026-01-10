-- Call events table for comprehensive voice flow logging
CREATE TABLE IF NOT EXISTS call_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    call_id uuid REFERENCES calls(id) ON DELETE CASCADE,
    event_type text NOT NULL,
    event_data jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Index for efficient call-based queries
CREATE INDEX idx_call_events_call_id ON call_events(call_id);

-- Index for time-based queries (recent events)
CREATE INDEX idx_call_events_created_at ON call_events(created_at DESC);

-- Index for filtering by event type
CREATE INDEX idx_call_events_type ON call_events(event_type);

-- Composite index for filtering by call + time
CREATE INDEX idx_call_events_call_created ON call_events(call_id, created_at DESC);
