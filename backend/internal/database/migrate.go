package database

import "fmt"

// RunMigrations creates all required tables if they don't already exist.
// Safe to call every time the server starts (IF NOT EXISTS guards it).
func RunMigrations(db *DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS dentists (
			dentist_id             SERIAL PRIMARY KEY,
			dentist_name           TEXT NOT NULL,
			image_base64           TEXT,
			chair_height           DOUBLE PRECISION NOT NULL DEFAULT 0,
			dentist_sex            TEXT,
			body_composition_data  TEXT,
			face_embedding         FLOAT8[],
			created_at             TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
		`ALTER TABLE dentists ADD COLUMN IF NOT EXISTS chair_height DOUBLE PRECISION NOT NULL DEFAULT 0`,
		`ALTER TABLE dentists ADD COLUMN IF NOT EXISTS image_path TEXT`,
		`ALTER TABLE dentists ADD COLUMN IF NOT EXISTS pose_landmarks JSONB`,
		`ALTER TABLE dentists ADD COLUMN IF NOT EXISTS ergonomics_sit_height NUMERIC`,
		`ALTER TABLE dentists ADD COLUMN IF NOT EXISTS update_at TIMESTAMPTZ NOT NULL DEFAULT now()`,
		`CREATE TABLE IF NOT EXISTS tb_face_embedding (
			id SERIAL PRIMARY KEY,
			dentist_id INTEGER NOT NULL REFERENCES dentists(dentist_id) ON DELETE CASCADE,
			embedding JSONB NOT NULL,
			model_version TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_face_embedding_dentist_id ON tb_face_embedding(dentist_id)`,

		`CREATE TABLE IF NOT EXISTS adjustable_items (
			adjust_item_id   SERIAL PRIMARY KEY,
			adjust_name      TEXT NOT NULL,
			data_used        TEXT,
			dentist_adjust   BOOLEAN NOT NULL DEFAULT FALSE,
			patient_adjust   BOOLEAN NOT NULL DEFAULT FALSE,
			created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
		`ALTER TABLE adjustable_items ADD COLUMN IF NOT EXISTS dentist_adjust BOOLEAN NOT NULL DEFAULT FALSE`,
		`ALTER TABLE adjustable_items ADD COLUMN IF NOT EXISTS patient_adjust BOOLEAN NOT NULL DEFAULT FALSE`,
		`INSERT INTO adjustable_items (adjust_item_id, adjust_name, data_used, dentist_adjust, patient_adjust)
		 VALUES (1, 'dentistchair', 'Adjust for a recognized dentist', TRUE, FALSE),
		        (2, 'clientchair', 'Adjust for a patient', FALSE, TRUE)
		 ON CONFLICT (adjust_item_id) DO UPDATE
		 SET adjust_name = EXCLUDED.adjust_name, data_used = EXCLUDED.data_used,
		     dentist_adjust = EXCLUDED.dentist_adjust, patient_adjust = EXCLUDED.patient_adjust`,
		`DO $$ BEGIN
		 IF pg_get_serial_sequence('adjustable_items', 'adjust_item_id') IS NOT NULL THEN
		   EXECUTE 'SELECT setval(pg_get_serial_sequence(''adjustable_items'', ''adjust_item_id''), GREATEST(COALESCE((SELECT MAX(adjust_item_id) FROM adjustable_items), 1), 1))';
		 END IF;
		 END $$`,

		`CREATE TABLE IF NOT EXISTS camera_history_logs (
			id                  SERIAL PRIMARY KEY,
			camera_name         TEXT NOT NULL DEFAULT 'Unknown Camera',
			dentist_id          TEXT NOT NULL,
			body_measurements   JSONB,
			adjustment_items    TEXT[],
			adjustment_status   TEXT,
			created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
		`ALTER TABLE camera_history_logs ADD COLUMN IF NOT EXISTS camera_name TEXT NOT NULL DEFAULT 'Unknown Camera'`,
		`ALTER TABLE camera_history_logs ADD COLUMN IF NOT EXISTS body_measurements JSONB`,
		`CREATE INDEX IF NOT EXISTS idx_camera_history_logs_created_at ON camera_history_logs(created_at)`,
		`CREATE TABLE IF NOT EXISTS mock3d (
			id SERIAL PRIMARY KEY,
			history_log_id INTEGER NOT NULL REFERENCES camera_history_logs(id) ON DELETE CASCADE,
			payload JSONB NOT NULL DEFAULT '{}'::jsonb,
			status BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
		`DO $$ BEGIN
			IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'mock3d' AND column_name = 'camera_log_id')
				AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'mock3d' AND column_name = 'history_log_id') THEN
				ALTER TABLE mock3d RENAME COLUMN camera_log_id TO history_log_id;
			END IF;
		END $$`,
		`CREATE INDEX IF NOT EXISTS idx_mock3d_history_log_id ON mock3d(history_log_id, created_at DESC)`,
	}

	for i, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("migration step %d failed: %w", i+1, err)
		}
	}
	return nil
}
