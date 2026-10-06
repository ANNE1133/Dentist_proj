package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"

	"backend/internal/domain"
)

// CreateDentistResult is what the handler maps into CreateDentistResponse.
type CreateDentistResult struct {
	Success   bool
	DentistID int
}

// RegisterDentistWithImage validates/extracts first, then stores the image and
// atomically attaches its object key and embedding to the dentist record.
func (s *Service) RegisterDentistWithImage(
	ctx context.Context,
	payload domain.Dentist,
	file multipart.File,
	filename string,
	contentType string,
) (*CreateDentistResult, error) {

	// Buffer once: MinIO needs a known size and the embedding extractor
	// needs the raw bytes too, and multipart.File can't be read twice.
	buf, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read image: %w", err)
	}
	if len(buf) == 0 {
		return nil, errors.New("empty image")
	}

	embedding, modelVersion, err := s.embedder.Extract(ctx, buf)
	if err != nil {
		return nil, fmt.Errorf("extract face embedding: %w", err)
	}
	if len(embedding) != 128 {
		return nil, fmt.Errorf("invalid face embedding: got %d dimensions, expected 128", len(embedding))
	}
	if modelVersion != "face_recognition_dlib_128d_v1" {
		return nil, fmt.Errorf("unsupported face embedding model %q", modelVersion)
	}

	created, err := s.db.CreateDentist(payload)
	if err != nil {
		return nil, fmt.Errorf("insert dentist: %w", err)
	}
	dentistID := created.DentistID
	objectName := fmt.Sprintf("dentists/%d/face%s", dentistID, filepath.Ext(filename))
	if err := s.minio.UploadTemporaryImage(ctx, bytes.NewReader(buf), objectName, int64(len(buf)), contentType); err != nil {
		s.cleanupFailedRegistration(objectName, dentistID)
		return nil, fmt.Errorf("upload image: %w", err)
	}
	if err := s.db.SaveRegistrationData(ctx, dentistID, objectName, domain.FaceEmbedding{DentistID: dentistID, Embedding: embedding, ModelVersion: modelVersion}); err != nil {
		s.cleanupFailedRegistration(objectName, dentistID)
		return nil, fmt.Errorf("save registration data: %w", err)
	}

	return &CreateDentistResult{Success: true, DentistID: dentistID}, nil
}

// cleanupFailedRegistration removes the orphaned MinIO object and, if the
// dentist row was already created, deletes it too — so a failure on one
// leg doesn't leave a half-registered dentist behind. Best-effort: errors
// here are swallowed since the caller is already returning the real error.
func (s *Service) cleanupFailedRegistration(objectName string, dentistID int) {
	_ = s.minio.DeleteObject(context.Background(), objectName)
	if dentistID != 0 {
		_ = s.db.DeleteDentist(context.Background(), dentistID)
	}
}

// --- existing base64 upload path, unchanged ---

func (s *Service) ListDentistRecords() ([]domain.Dentist, error) {
	return s.db.ListDentists()
}

func (s *Service) GetDentistRecordByID(id int) (domain.Dentist, error) {
	if id <= 0 {
		return domain.Dentist{}, errors.New("invalid dentist id")
	}
	return s.db.GetDentistByID(id)
}
