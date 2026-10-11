package service

import (
	"backend/internal/domain"
	"errors"
	"fmt"
	"log"
	"math"
)

// DetectedFace is what the live-camera endpoint returns per face found,
// ready for the frontend to draw a bounding box + label.
type DetectedFace struct {
	BBox      [4]int  `json:"bbox"` // x, y, w, h in pixels
	Label     string  `json:"label"`
	DentistID string  `json:"dentistId,omitempty"`
	Matched   bool    `json:"matched"`
	Distance  float64 `json:"distance"`
}

// DetectFaces takes one camera frame, asks the python service for faces +
// embeddings, then matches each embedding against registered dentists.
func (s *Service) DetectFaces(imageBase64 string) ([]DetectedFace, error) {
	return s.DetectFacesWithPose(imageBase64, nil)
}

func (s *Service) DetectFacesWithPose(imageBase64 string, pose []domain.PoseLandmark) ([]DetectedFace, error) {
	return s.DetectFacesWithPoseAndCamera(imageBase64, pose, "Unknown Camera")
}

func (s *Service) DetectFacesWithPoseAndCamera(imageBase64 string, pose []domain.PoseLandmark, cameraName string) ([]DetectedFace, error) {
	faces, err := s.python.Detect(imageBase64)
	if err != nil {
		return nil, fmt.Errorf("detect faces: %w", err)
	}

	dentists, err := s.db.ListDentistsForMatching()
	if err != nil {
		return nil, fmt.Errorf("load dentists for matching: %w", err)
	}

	out := make([]DetectedFace, 0, len(faces))
	measurements := BodyMeasurementsFromPose(pose)
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
		dentistID := "UNKNOWN"
		if ok {
			label = match.DentistName
			dentistID = fmt.Sprint(match.DentistID)
		}
		if err := s.LogDetectedPerson(dentistID, measurements, cameraName); err != nil {
			// Keep live recognition available if history storage is temporarily down.
			log.Printf("persist camera history for %s: %v", dentistID, err)
		}
		out = append(out, DetectedFace{
			BBox:      f.BBox,
			Label:     label,
			DentistID: dentistID,
			Matched:   ok,
			Distance:  dist,
		})
	}
	return out, nil
}

// BodyMeasurementsFromPose calculates 2D keypoint distances from the camera's
// pose output. Coordinates are normalized, so only the unit is marked mock;
// measurements with missing or low-visibility landmarks are omitted. If no
// complete measurements can be calculated, nil is stored as SQL/JSON null.
func BodyMeasurementsFromPose(pose []domain.PoseLandmark) []domain.BodyMeasurement {
	const minVisibility = 0.5
	pointUsable := func(index int) bool {
		if index >= len(pose) {
			return false
		}
		p := pose[index]
		return p.Visibility >= minVisibility && !math.IsNaN(p.X) && !math.IsNaN(p.Y) && !math.IsInf(p.X, 0) && !math.IsInf(p.Y, 0)
	}
	makeMeasurement := func(part string, indices ...int) (domain.BodyMeasurement, bool) {
		for _, index := range indices {
			if !pointUsable(index) {
				return domain.BodyMeasurement{}, false
			}
		}
		var value float64
		confidence := 1.0
		for i, index := range indices {
			point := pose[index]
			confidence = math.Min(confidence, point.Visibility)
			if i > 0 {
				previous := pose[indices[i-1]]
				value += math.Hypot(point.X-previous.X, point.Y-previous.Y)
			}
		}
		return domain.BodyMeasurement{
			Part: part, Value: math.Round(value*10000) / 10000, Unit: "mock_units",
			Confidence: &confidence, Method: "normalized_pose_2d_keypoint_distance",
		}, true
	}

	measurements := make([]domain.BodyMeasurement, 0, 5)
	for _, spec := range []struct {
		part    string
		indices []int
	}{
		{"left_leg_length", []int{23, 25, 27}},
		{"right_leg_length", []int{24, 26, 28}},
		{"left_arm_length", []int{11, 13, 15}},
		{"right_arm_length", []int{12, 14, 16}},
	} {
		if measurement, ok := makeMeasurement(spec.part, spec.indices...); ok {
			measurements = append(measurements, measurement)
		}
	}
	if pointUsable(11) && pointUsable(12) && pointUsable(23) && pointUsable(24) {
		shoulderMidpoint := domain.PoseLandmark{X: (pose[11].X + pose[12].X) / 2, Y: (pose[11].Y + pose[12].Y) / 2}
		hipMidpoint := domain.PoseLandmark{X: (pose[23].X + pose[24].X) / 2, Y: (pose[23].Y + pose[24].Y) / 2}
		confidence := math.Min(math.Min(pose[11].Visibility, pose[12].Visibility), math.Min(pose[23].Visibility, pose[24].Visibility))
		measurements = append(measurements, domain.BodyMeasurement{
			Part: "torso_length", Value: math.Round(math.Hypot(hipMidpoint.X-shoulderMidpoint.X, hipMidpoint.Y-shoulderMidpoint.Y)*10000) / 10000,
			Unit: "mock_units", Confidence: &confidence, Method: "normalized_pose_2d_keypoint_distance",
		})
	}
	if len(measurements) == 0 {
		return nil
	}
	return measurements
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
