package handlers

import (
	"encoding/json"
	"net/http"
)

// GetCameraLogs handles GET /api/camera/logs to fetch all camera history logs. 
func (h *Handler) GetCameraLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	logs, err := h.service.ListCameraLogs()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(logs)
}

// CreateCameraLogs handles POST /api/camera/logs for the existing log analysis flow.
func (h *Handler) CreateCameraLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var payload struct {
		DentistName string  `json:"dentistName"`
		Label       string  `json:"label"`
		Confidence  float64 `json:"confidence"`
		Details     string  `json:"details"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}