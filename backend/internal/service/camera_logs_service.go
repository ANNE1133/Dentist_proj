package service

import (
	"errors"

	"backend/internal/domain"
)

func (s *Service) CreateCameraLog(l domain.CameraLogs) (domain.CameraLogs, error) {
	if l.DentistID == "" {
		return domain.CameraLogs{}, errors.New("dentistId is required")
	}
	return s.db.CreateCameraLog(l)
}

func (s *Service) ListCameraLogs() ([]domain.CameraLogs, error) {
	return s.db.ListCameraLogs()
}
