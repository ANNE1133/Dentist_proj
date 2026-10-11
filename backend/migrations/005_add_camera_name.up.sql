ALTER TABLE camera_history_logs
    ADD COLUMN IF NOT EXISTS camera_name TEXT NOT NULL DEFAULT 'Unknown Camera';
