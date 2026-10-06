CREATE TABLE camera_history_logs (
    id SERIAL PRIMARY KEY,
    dentist_id VARCHAR(100) NOT NULL,
    body_composition TEXT,
    adjustment_items TEXT[],
    adjustment_status VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);