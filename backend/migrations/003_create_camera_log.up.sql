CREATE TABLE camera_history_logs (
    id SERIAL PRIMARY KEY,
    dentist_id VARCHAR(100) NOT NULL,
    adjustment_items TEXT[],
    adjustment_status VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_camera_history_logs_created_at ON camera_history_logs(created_at);
