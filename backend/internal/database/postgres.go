package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/lib/pq"

	"backend/internal/config"
	"backend/internal/domain"
)

// DB wraps a *sql.DB so we can attach query methods to it.
type DB struct {
	*sql.DB
}

type DentistStore interface {
	CreateDentist(d domain.Dentist) (domain.Dentist, error)
	ListDentists() ([]domain.Dentist, error)
	GetDentistByID(id int) (domain.Dentist, error)
	CreateFaceEmbedding(ctx context.Context, e domain.FaceEmbedding) error
	DeleteDentist(ctx context.Context, id int) error
}

// Connect opens a connection pool to Postgres and verifies it with a ping.
func Connect(cfg config.Config) (*DB, error) {
	sqlDB, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return &DB{sqlDB}, nil
}

// ---------- Dentists ----------

func (db *DB) GetDentistByID(id int) (domain.Dentist, error) {
	const q = `
		SELECT dentist_id, dentist_name, COALESCE(image_path, ''), created_at
		FROM dentists
		WHERE dentist_id = $1`

	var d domain.Dentist
	err := db.QueryRow(q, id).Scan(&d.DentistID, &d.DentistName, &d.ImagePath, &d.CreatedAt)
	if err != nil {
		return domain.Dentist{}, fmt.Errorf("get dentist %d: %w", id, err)
	}
	return d, nil
}

func (db *DB) ListDentists() ([]domain.Dentist, error) {
	const q = `
		SELECT dentist_id, dentist_name, COALESCE(image_path, ''), created_at
		FROM dentists
		ORDER BY dentist_id DESC`

	rows, err := db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("list dentists: %w", err)
	}
	defer rows.Close()

	var out []domain.Dentist
	for rows.Next() {
		var d domain.Dentist
		if err := rows.Scan(&d.DentistID, &d.DentistName, &d.ImagePath, &d.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan dentist: %w", err)
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// ListDentistsForMatching loads one current embedding per dentist. New registrations
// create one embedding; DISTINCT ON also handles legacy duplicate rows safely.
func (db *DB) ListDentistsForMatching() ([]domain.Dentist, error) {
	const q = `SELECT d.dentist_id, d.dentist_name, COALESCE(d.image_path, ''),
		COALESCE(e.embedding, 'null'::jsonb), COALESCE(e.model_version, '')
		FROM dentists d LEFT JOIN LATERAL (
			SELECT embedding, model_version FROM tb_face_embedding
			WHERE dentist_id = d.dentist_id ORDER BY created_at DESC, id DESC LIMIT 1
		) e ON true ORDER BY d.dentist_id`
	rows, err := db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("list dentists with embeddings: %w", err)
	}
	defer rows.Close()
	var out []domain.Dentist
	for rows.Next() {
		var d domain.Dentist
		var raw []byte
		if err := rows.Scan(&d.DentistID, &d.DentistName, &d.ImagePath, &raw, &d.ModelVersion); err != nil {
			return nil, fmt.Errorf("scan dentist embedding: %w", err)
		}
		if len(raw) > 0 && string(raw) != "null" {
			if err := json.Unmarshal(raw, &d.FaceEmbedding); err != nil {
				return nil, fmt.Errorf("decode dentist embedding %d: %w", d.DentistID, err)
			}
			if len(d.FaceEmbedding) != 128 {
				return nil, fmt.Errorf("dentist %d embedding has %d dimensions, expected 128", d.DentistID, len(d.FaceEmbedding))
			}
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// CreateDentist inserts a new row into dentists. Used both by the legacy
// base64 upload path (CreateDentistRecord) and the new MinIO + face
// embedding flow (RegisterDentistWithImage) — the latter only fills
// DentistName + ImagePath, leaving DentistSex/BodyCompositionData empty,
// so those columns need to allow NULL/default.
func (db *DB) CreateDentist(d domain.Dentist) (domain.Dentist, error) {
	const q = `
		INSERT INTO dentists (dentist_name, image_path, chair_height)
		VALUES ($1, NULL, 0)
		RETURNING dentist_id, created_at`

	err := db.QueryRow(q, d.DentistName).
		Scan(&d.DentistID, &d.CreatedAt)
	if err != nil {
		return domain.Dentist{}, fmt.Errorf("insert dentist: %w", err)
	}
	return d, nil
}

// SaveRegistrationData atomically persists the private MinIO object key and
// its compatible face embedding after the image upload succeeds.
func (db *DB) SaveRegistrationData(ctx context.Context, id int, imagePath string, embedding domain.FaceEmbedding) error {
	if len(embedding.Embedding) != 128 {
		return fmt.Errorf("embedding has %d dimensions, expected 128", len(embedding.Embedding))
	}
	if embedding.ModelVersion == "" {
		return fmt.Errorf("embedding model version is required")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `UPDATE dentists SET image_path=$2 WHERE dentist_id=$1`, id, imagePath); err != nil {
		return fmt.Errorf("save dentist image path: %w", err)
	}
	data, err := json.Marshal(embedding.Embedding)
	if err != nil {
		return fmt.Errorf("marshal embedding: %w", err)
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO tb_face_embedding (dentist_id, embedding, model_version) VALUES ($1,$2,$3)`, id, string(data), embedding.ModelVersion); err != nil {
		return fmt.Errorf("insert face embedding: %w", err)
	}
	return tx.Commit()
}

// DeleteDentist is used to roll back a dentist row if the paired face
// embedding insert fails (RegisterDentistWithImage runs the dentist insert
// and the embedding extraction concurrently, so they can't share one
// transaction — this is the manual undo for that).
func (db *DB) DeleteDentist(ctx context.Context, id int) error {
	const q = `DELETE FROM dentists WHERE dentist_id = $1`
	_, err := db.ExecContext(ctx, q, id)
	return err
}

// ---------- Face Embeddings ----------

// CreateFaceEmbedding stores the embedding as a JSON array of float32 in a
// JSONB column. If you install the pgvector extension, change the embedding
// column to `vector(N)` and use github.com/pgvector/pgvector-go instead of
// json.Marshal so you get real similarity search (<-> operator).
func (db *DB) CreateFaceEmbedding(ctx context.Context, e domain.FaceEmbedding) error {
	if len(e.Embedding) != 128 {
		return fmt.Errorf("embedding has %d dimensions, expected 128", len(e.Embedding))
	}
	if e.ModelVersion != "face_recognition_dlib_128d_v1" {
		return fmt.Errorf("unsupported embedding model %q", e.ModelVersion)
	}
	const q = `
		INSERT INTO tb_face_embedding (dentist_id, embedding, model_version)
		VALUES ($1, $2, $3)`

	emb, err := json.Marshal(e.Embedding)
	if err != nil {
		return fmt.Errorf("marshal embedding: %w", err)
	}
	if _, err := db.ExecContext(ctx, q, e.DentistID, string(emb), e.ModelVersion); err != nil {
		return fmt.Errorf("insert face embedding: %w", err)
	}
	return nil
}

// ---------- Adjustable Items ----------

func (db *DB) CreateAdjustItem(a domain.AdjustableItem) (domain.AdjustableItem, error) {
	const q = `
		INSERT INTO adjustable_items (adjust_item_id, adjust_name, data_used)
		VALUES ($1, $2, $3)
		RETURNING adjust_item_id, created_at`

	err := db.QueryRow(q, a.AdjustItemID, a.AdjustName, a.DataUsed).Scan(&a.AdjustItemID, &a.CreatedAt)
	if err != nil {
		return domain.AdjustableItem{}, fmt.Errorf("insert adjust item: %w", err)
	}
	return a, nil
}

func (db *DB) GetAdjustItemByID(id int) (domain.AdjustableItem, error) {
	const q = `
		SELECT adjust_item_id, adjust_name, data_used, created_at
		FROM adjustable_items
		WHERE adjust_item_id = $1`

	var a domain.AdjustableItem
	err := db.QueryRow(q, id).Scan(&a.AdjustItemID, &a.AdjustName, &a.DataUsed, &a.CreatedAt)
	if err != nil {
		return domain.AdjustableItem{}, fmt.Errorf("get adjust item %d: %w", id, err)
	}
	return a, nil
}

func (db *DB) ListAdjustItems() ([]domain.AdjustableItem, error) {
	const q = `
		SELECT adjust_item_id, adjust_name, data_used, created_at
		FROM adjustable_items
		ORDER BY adjust_item_id DESC`

	rows, err := db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("list adjust items: %w", err)
	}
	defer rows.Close()

	var out []domain.AdjustableItem
	for rows.Next() {
		var a domain.AdjustableItem
		if err := rows.Scan(&a.AdjustItemID, &a.AdjustName, &a.DataUsed, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan adjust item: %w", err)
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// ---------- Camera History Logs ----------

func (db *DB) CreateCameraLog(l domain.CameraLogs) (domain.CameraLogs, error) {
	const q = `
		INSERT INTO camera_logs (dentist_id, body_composition, adjustment_items, adjustment_status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`

	err := db.QueryRow(q,
		l.DentistID, l.BodyComposition, pq.Array(l.AdjustmentItems), l.AdjustmentStatus,
	).Scan(&l.ID, &l.CreatedAt)
	if err != nil {
		return domain.CameraLogs{}, fmt.Errorf("insert camera log: %w", err)
	}
	return l, nil
}

func (db *DB) ListCameraLogs() ([]domain.CameraLogs, error) {
	const q = `
		SELECT id, dentist_id, body_composition, adjustment_items, adjustment_status, created_at
		FROM camera_logs
		ORDER BY created_at DESC`

	rows, err := db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("list camera history logs: %w", err)
	}
	defer rows.Close()

	var out []domain.CameraLogs
	for rows.Next() {
		var l domain.CameraLogs
		if err := rows.Scan(
			&l.ID, &l.DentistID, &l.BodyComposition, pq.Array(&l.AdjustmentItems), &l.AdjustmentStatus, &l.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan cameraa log: %w", err)
		}
		out = append(out, l)
	}
	return out, rows.Err()
}
