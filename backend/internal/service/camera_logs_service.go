package service

import (
	"errors"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"backend/internal/domain"
)

var cameraLogMu sync.Mutex
var cameraLogLastSeen = map[string]time.Time{}

func (s *Service) CreateCameraLog(l domain.CameraLogs) (domain.CameraLogs, error) {
	l.DentistID = strings.TrimSpace(l.DentistID)
	l.CameraName = strings.TrimSpace(l.CameraName)
	if l.CameraName == "" {
		l.CameraName = "Unknown Camera"
	}
	if l.DentistID == "" {
		return domain.CameraLogs{}, errors.New("dentistId is required")
	}
	for i, measurement := range l.BodyMeasurements {
		measurement.Part = strings.TrimSpace(measurement.Part)
		measurement.Unit = strings.TrimSpace(measurement.Unit)
		measurement.Method = strings.TrimSpace(measurement.Method)
		if measurement.Part == "" || measurement.Unit == "" || measurement.Value < 0 || math.IsNaN(measurement.Value) || math.IsInf(measurement.Value, 0) {
			return domain.CameraLogs{}, errors.New("each body measurement requires a part, a unit, and a non-negative value")
		}
		if measurement.Confidence != nil && (*measurement.Confidence < 0 || *measurement.Confidence > 1) {
			return domain.CameraLogs{}, errors.New("measurement confidence must be between 0 and 1")
		}
		l.BodyMeasurements[i] = measurement
	}
	return s.db.CreateCameraLog(l)
}

func (s *Service) ListCameraLogs() ([]domain.CameraLogs, error) {
	return s.db.ListCameraLogs()
}

func (s *Service) ListCameraLogsPage(page, limit int, start, end *time.Time) (domain.CameraLogsPage, error) {
	return s.db.ListCameraLogsPage(page, limit, start, end)
}

// LogDetectedPerson writes at most one row per identity during the configured cooldown.
func (s *Service) LogDetectedPerson(dentistID string, measurements []domain.BodyMeasurement, cameraName string) error {
	identity := strings.TrimSpace(dentistID)
	if identity == "" {
		return errors.New("dentistId is required")
	}
	cameraName = strings.TrimSpace(cameraName)
	if cameraName == "" {
		cameraName = "Unknown Camera"
	}
	cooldown := 60 * time.Second
	if raw := os.Getenv("CAMERA_LOG_COOLDOWN_SECONDS"); raw != "" {
		seconds, err := strconv.Atoi(raw)
		if err != nil || seconds < 0 {
			return errors.New("CAMERA_LOG_COOLDOWN_SECONDS must be a non-negative integer")
		}
		cooldown = time.Duration(seconds) * time.Second
	}
	now := time.Now()
	cameraLogMu.Lock()
	logKey := identity + "|" + cameraName
	last, seen := cameraLogLastSeen[logKey]
	if seen && cooldown > 0 && now.Sub(last) < cooldown {
		cameraLogMu.Unlock()
		return nil
	}
	// Reserve before insertion so concurrent frames cannot create duplicates.
	cameraLogLastSeen[logKey] = now
	cameraLogMu.Unlock()
	status := "not_applicable"
	adjustableItems, err := s.db.ListAdjustItems()
	if err != nil {
		cameraLogMu.Lock()
		if cameraLogLastSeen[logKey] == now {
			delete(cameraLogLastSeen, logKey)
		}
		cameraLogMu.Unlock()
		return err
	}
	recognizedDentist := identity != "UNKNOWN"
	adjustmentItems := make([]string, 0)
	for _, item := range adjustableItems {
		applies := item.PatientAdjust
		if recognizedDentist {
			applies = item.DentistAdjust
		}
		if applies {
			adjustmentItems = append(adjustmentItems, item.AdjustName)
		}
	}
	if len(measurements) > 0 && len(adjustmentItems) > 0 {
		status = "pending"
	}
	if _, err := s.CreateCameraLog(domain.CameraLogs{
		CameraName: cameraName, DentistID: identity, BodyMeasurements: measurements,
		AdjustmentItems: adjustmentItems, AdjustmentStatus: status,
	}); err != nil {
		cameraLogMu.Lock()
		if cameraLogLastSeen[logKey] == now {
			delete(cameraLogLastSeen, logKey)
		}
		cameraLogMu.Unlock()
		return err
	}
	return nil
}

func (s *Service) Save3DAdjustment(adjustment domain.Adjustment3D) error {
	if adjustment.HistoryLogID < 1 {
		return errors.New("historyLogId must be a positive integer")
	}
	return s.db.Create3DAdjustment(adjustment)
}
