package domain

import "time"

type Dentist struct {
	DentistID           int       `json:"dentistId" db:"dentist_id"`
	DentistName         string    `json:"dentistName" db:"dentist_name"`
	FaceEmbedding       []float64 `json:"faceEmbedding,omitempty"`
	ModelVersion        string    `json:"modelVersion,omitempty"`
	ImagePath           string    `json:"imagePath,omitempty" db:"image_path"`
	PoseLandmarks       []byte    `json:"poseLandmarks,omitempty" db:"pose_landmarks"` // raw JSONB, set by a later step
	ErgonomicsSitHeight *float64  `json:"ergonomicsSitHeight,omitempty" db:"ergonomics_sit_height"`
	CreatedAt           time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt           time.Time `json:"updatedAt,omitempty" db:"update_at"`
}

// FaceEmbedding mirrors the new tb_face_embedding table.
type FaceEmbedding struct {
	ID           int       `json:"id" db:"id"`
	DentistID    int       `json:"dentistId" db:"dentist_id"`
	Embedding    []float64 `json:"embedding" db:"embedding"`
	ModelVersion string    `json:"modelVersion" db:"model_version"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
}

type AdjustableItem struct {
	AdjustItemID string `json:"adjustItemId"`
	AdjustName   string `json:"adjustName"`
	DataUsed     string `json:"dataUsed"`
	CreatedAt    string `json:"created_at"`
}

type CameraLogs struct {
	ID               int      `json:"id"`
	DentistID        string   `json:"dentistId"`
	BodyComposition  string   `json:"bodyComposition"`
	AdjustmentItems  []string `json:"adjustmentItems"`
	AdjustmentStatus string   `json:"adjustmentStatus"`
	CreatedAt        string   `json:"created_at"`
}

type AnalysisResult struct {
	Label       string  `json:"label"`
	Confidence  float64 `json:"confidence"`
	Details     string  `json:"details"`
	ChairHeight float64 `json:"chair_height"`
}
