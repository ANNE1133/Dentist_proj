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
			created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,

		`CREATE TABLE IF NOT EXISTS camera_logs (
			id                  SERIAL PRIMARY KEY,
			dentist_id          TEXT NOT NULL,
			body_composition    TEXT,
			adjustment_items    TEXT[],
			adjustment_status   TEXT,
			created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
	}

	for i, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("migration step %d failed: %w", i+1, err)
		}
	}
	return nil
}
