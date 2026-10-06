CREATE TABLE dentists (
    dentist_id SERIAL PRIMARY KEY,
    image_base64 TEXT,
    chair_height DOUBLE PRECISION NOT NULL,
    dentist_name VARCHAR(255),
    body_composition_data TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);