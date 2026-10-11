CREATE TABLE adjustable_items (
    id SERIAL PRIMARY KEY,
    adjust_item_id VARCHAR(100) UNIQUE NOT NULL,
    adjust_name VARCHAR(255) NOT NULL,
    data_used TEXT,
    dentist_adjust BOOLEAN NOT NULL DEFAULT FALSE,
    patient_adjust BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO adjustable_items (adjust_item_id, adjust_name, data_used, dentist_adjust, patient_adjust)
VALUES ('1', 'dentistchair', 'Adjust for a recognized dentist', TRUE, FALSE),
       ('2', 'clientchair', 'Adjust for a patient', FALSE, TRUE)
ON CONFLICT (adjust_item_id) DO UPDATE
SET adjust_name = EXCLUDED.adjust_name, data_used = EXCLUDED.data_used,
    dentist_adjust = EXCLUDED.dentist_adjust, patient_adjust = EXCLUDED.patient_adjust;
