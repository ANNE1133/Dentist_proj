CREATE TABLE adjustable_items (
    id SERIAL PRIMARY KEY,
    adjust_item_id VARCHAR(100) UNIQUE NOT NULL,
    adjust_name VARCHAR(255) NOT NULL,
    data_used TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);