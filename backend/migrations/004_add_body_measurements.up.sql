-- Body composition is not inferred from pose landmarks. Store explicitly
-- measured or estimated per-part values as structured JSON instead.
ALTER TABLE camera_history_logs
    ADD COLUMN IF NOT EXISTS body_measurements JSONB;
