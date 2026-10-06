package service

import (
	"errors"
	"backend/internal/domain"
)

// ---------- Adjustable Items ----------

func (s *Service) CreateAdjustItem(a domain.AdjustableItem) (domain.AdjustableItem, error) {
	if a.AdjustName == "" {
		return domain.AdjustableItem{}, errors.New("adjustName is required")
	}
	return s.db.CreateAdjustItem(a)
}

func (s *Service) ListAdjustItems() ([]domain.AdjustableItem, error) {
	return s.db.ListAdjustItems()
}

func (s *Service) GetAdjustItemByID(id int) (domain.AdjustableItem, error) {
	if id <= 0 {
		return domain.AdjustableItem{}, errors.New("invalid adjust item id")
	}
	return s.db.GetAdjustItemByID(id)
}
