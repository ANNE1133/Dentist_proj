-- Adds the columns needed for the MinIO + face-embedding flow to the
-- existing `dentists` table, and creates the new tb_face_embedding table.
-- Written as ALTER, not CREATE, because `dentists` already exists and
-- camera_logs has a foreign key into it — renaming it is out of scope here.

ALTER TABLE dentists
    ADD COLUMN IF NOT EXISTS image_path            TEXT,
    ADD COLUMN IF NOT EXISTS pose_landmarks         JSONB,
    ADD COLUMN IF NOT EXISTS ergonomics_sit_height  NUMERIC,
    ADD COLUMN IF NOT EXISTS update_at              TIMESTAMPTZ NOT NULL DEFAULT now();

CREATE TABLE IF NOT EXISTS tb_face_embedding (
    id             SERIAL PRIMARY KEY,
    dentist_id     INTEGER NOT NULL REFERENCES dentists(dentist_id) ON DELETE CASCADE,
    embedding      JSONB NOT NULL,     -- swap to `vector(N)` if you install the pgvector extension
    model_version  TEXT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_face_embedding_dentist_id ON tb_face_embedding(dentist_id);

-- Optional, for real similarity search instead of JSONB:
-- CREATE EXTENSION IF NOT EXISTS vector;
-- ALTER TABLE tb_face_embedding ALTER COLUMN embedding TYPE vector(512) USING embedding::text::vector;
-- CREATE INDEX ON tb_face_embedding USING ivfflat (embedding vector_cosine_ops);