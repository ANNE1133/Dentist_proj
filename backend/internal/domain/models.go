package domain

import (
	"encoding/json"
	"time"
)

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
	AdjustItemID  string `json:"adjustItemId"`
	AdjustName    string `json:"adjustName"`
	DataUsed      string `json:"dataUsed"`
	DentistAdjust bool   `json:"dentistAdjust"`
	PatientAdjust bool   `json:"patientAdjust"`
	CreatedAt     string `json:"created_at"`
}

type CameraLogs struct {
	ID               int               `json:"id"`
	CameraName       string            `json:"cameraName"`
	DentistID        string            `json:"dentistId"`
	BodyMeasurements []BodyMeasurement `json:"bodyMeasurements"`
	AdjustmentItems  []string          `json:"adjustmentItems"`
	AdjustmentStatus string            `json:"adjustmentStatus"`
	CreatedAt        string            `json:"created_at"`
}

// Adjustment3D stores the payload and completion state returned by the 3D adjustment flow.
type Adjustment3D struct {
	HistoryLogID    int                `json:"historyLogId"`
	Status          bool               `json:"status"`
	Data            json.RawMessage    `json:"data,omitempty"`
	AdjustmentItems []Adjustment3DItem `json:"adjustmentItems,omitempty"`
}

type Adjustment3DItem struct {
	AdjustID   string          `json:"adjustId"`
	AdjustName string          `json:"adjustName"`
	DataUsed   json.RawMessage `json:"dataUsed"`
}

// BodyMeasurement stores an estimated measurement for one body part. Value and
// unit describe the measurement; Method identifies how it was obtained.
type BodyMeasurement struct {
	Part       string   `json:"part"`
	Value      float64  `json:"value"`
	Unit       string   `json:"unit"`
	Confidence *float64 `json:"confidence,omitempty"`
	Method     string   `json:"method,omitempty"`
}

// PoseLandmark is a normalized MediaPipe pose point from a camera frame.
type PoseLandmark struct {
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	Z          float64 `json:"z"`
	Visibility float64 `json:"visibility"`
}

type CameraLogsPage struct {
	Data       []CameraLogs `json:"data"`
	Page       int          `json:"page"`
	Limit      int          `json:"limit"`
	Total      int          `json:"total"`
	TotalPages int          `json:"totalPages"`
}

type AnalysisResult struct {
	Label       string  `json:"label"`
	Confidence  float64 `json:"confidence"`
	Details     string  `json:"details"`
	ChairHeight float64 `json:"chair_height"`
}
