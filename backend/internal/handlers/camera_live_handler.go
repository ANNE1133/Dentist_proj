package handlers

import (
	"encoding/json"
	"net/http"
)

// GetCameraLive handles POST /api/camera/live.
// The iPhone camera (frontend) captures a frame, base64-encodes it, and POSTs
// it here. The handler asks the service to detect faces in the frame and
// match them against registered dentists, returning bounding boxes + labels
// for the frontend to draw as an overlay.
func (h *Handler) GetCameraLive(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		ImageBase64 string `json:"imageBase64"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if payload.ImageBase64 == "" {
		http.Error(w, "imageBase64 is required", http.StatusBadRequest)
		return
	}

	faces, err := h.service.DetectFaces(payload.ImageBase64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"faces": faces})
}