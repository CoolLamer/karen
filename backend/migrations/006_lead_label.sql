-- Add lead_label column to classify calls by business priority
ALTER TABLE call_screening_results
ADD COLUMN IF NOT EXISTS lead_label text NOT NULL DEFAULT 'nezjisteno';
