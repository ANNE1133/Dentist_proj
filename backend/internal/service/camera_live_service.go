package service

import (
	"backend/internal/domain"
	"errors"
	"fmt"
)

// DetectedFace is what the live-camera endpoint returns per face found,
// ready for the frontend to draw a bounding box + label.
type DetectedFace struct {
	BBox     [4]int  `json:"bbox"` // x, y, w, h in pixels
	Label    string  `json:"label"`
	Matched  bool    `json:"matched"`
	Distance float64 `json:"distance"`
}

// DetectFaces takes one camera frame, asks the python service for faces +
// embeddings, then matches each embedding against registered dentists.
func (s *Service) DetectFaces(imageBase64 string) ([]DetectedFace, error) {
	faces, err := s.python.Detect(imageBase64)
	if err != nil {
		return nil, fmt.Errorf("detect faces: %w", err)
	}

	dentists, err := s.db.ListDentistsForMatching()
	if err != nil {
		return nil, fmt.Errorf("load dentists for matching: %w", err)
	}

	out := make([]DetectedFace, 0, len(faces))
	for _, f := range faces {
		if len(f.Embedding) != 128 {
			return nil, fmt.Errorf("live face embedding has %d dimensions, expected 128", len(f.Embedding))
		}
		compatible := make([]domain.Dentist, 0, len(dentists))
		for _, dentist := range dentists {
			if len(dentist.FaceEmbedding) == 128 && dentist.ModelVersion == "face_recognition_dlib_128d_v1" {
				compatible = append(compatible, dentist)
			}
		}
		match, dist, ok := matchEmbedding(f.Embedding, compatible)
		label := "unknown"
		if ok {
			label = match.DentistName
			if label == "" {
				label = match.DentistName
			}
		}
		out = append(out, DetectedFace{
			BBox:     f.BBox,
			Label:    label,
			Matched:  ok,
			Distance: dist,
		})
	}
	return out, nil
}

// RegisterDentistWithFace runs the registration photo through the python
// service (face embedding + estimated chair height), then persists the
// dentist record with both filled in.
func (s *Service) RegisterDentistWithFace(d domain.Dentist) (domain.Dentist, error) {
	if d.DentistName == "" {
		return domain.Dentist{}, errors.New("name is required")
	}

	result, err := s.python.Register(d.DentistName, d.DentistName)
	if err != nil {
		return domain.Dentist{}, fmt.Errorf("register face: %w", err)
	}

	d.FaceEmbedding = result.FaceEmbedding
	return s.db.CreateDentist(d)
}

// AnalyzeImage is a placeholder for the real analysis pipeline.
// TODO: replace with actual model/vision-service call. Currently returns a
// mock result and does NOT persist a log — wire this to CreateCameraHistoryLog
// once the real analysis payload (body composition, adjustment items) is known.
func (s *Service) AnalyzeImage(dentistName string) (domain.AnalysisResult, error) {
	if dentistName == "" {
		return domain.AnalysisResult{}, errors.New("dentistName is required")
	}
	return domain.AnalysisResult{
		Label:      "unanalyzed",
		Confidence: 0,
		Details:    fmt.Sprintf("no analysis pipeline wired yet for %s", dentistName),
	}, nil
}
