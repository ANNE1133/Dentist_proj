package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"backend/internal/domain"
)

// GetCameraLogs handles GET /api/camera/logs to fetch all camera history logs.
func (h *Handler) GetCameraLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	page, limit := 1, 20
	var err error
	if raw := r.URL.Query().Get("page"); raw != "" {
		page, err = strconv.Atoi(raw)
		if err != nil || page < 1 {
			http.Error(w, "page must be a positive integer", http.StatusBadRequest)
			return
		}
	}
	if raw := r.URL.Query().Get("limit"); raw != "" {
		limit, err = strconv.Atoi(raw)
		if err != nil || limit < 1 || limit > 100 {
			http.Error(w, "limit must be between 1 and 100", http.StatusBadRequest)
			return
		}
	}
	start, err := parseLogTime(r.URL.Query().Get("startTime"))
	if err != nil {
		http.Error(w, "invalid startTime", http.StatusBadRequest)
		return
	}
	end, err := parseLogTime(r.URL.Query().Get("endTime"))
	if err != nil {
		http.Error(w, "invalid endTime", http.StatusBadRequest)
		return
	}
	if start != nil && end != nil && start.After(*end) {
		http.Error(w, "startTime must be before or equal to endTime", http.StatusBadRequest)
		return
	}
	logs, err := h.service.ListCameraLogsPage(page, limit, start, end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(logs)
}

func parseLogTime(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02T15:04:05", "2006-01-02T15:04"} {
		if t, err := time.Parse(layout, raw); err == nil {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("invalid time")
}

// CreateCameraLogs handles POST /api/camera/logs for the existing log analysis flow.
func (h *Handler) CreateCameraLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var payload domain.CameraLogs
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.DentistID) == "" {
		http.Error(w, "dentistId is required", http.StatusBadRequest)
		return
	}
	log, err := h.service.CreateCameraLog(payload)
	if err != nil {
		http.Error(w, "failed to save camera history", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(log)
}

// Create3DAdjustment handles POST /api/3dadust for a camera history record.
func (h *Handler) Create3DAdjustment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var payload domain.Adjustment3D
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if payload.HistoryLogID < 1 {
		http.Error(w, "historyLogId must be a positive integer", http.StatusBadRequest)
		return
	}
	if err := h.service.Save3DAdjustment(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{"message": "3D adjustment created"})
}
