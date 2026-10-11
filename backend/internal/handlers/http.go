package handlers

import (
	"net/http"

	"backend/internal/service"
)

type Handler struct {
	service *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

// RegisterRoutes wires HTTP method + path -> handler.
// NOTE: requires Go 1.22+ for method-prefixed mux patterns ("GET /path").
// Check your go.mod `go` directive - if it's below 1.22, tell me and I'll
// rewrite this using a method switch inside each handler instead.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.Health)

	mux.HandleFunc("GET /api/dentist", h.GetAllDentists)
	mux.HandleFunc("POST /api/dentist", h.CreateDentist)
	mux.HandleFunc("GET /api/dentist/{id}", h.GetDentist)

	mux.HandleFunc("GET /api/adjustitem", h.GetAllAdjustableItems)
	mux.HandleFunc("POST /api/adjustitem", h.CreateAdjustableItem)
	mux.HandleFunc("GET /api/adjustitem/{id}", h.GetAdjustableItembyID)

	mux.HandleFunc("GET /api/camera/logs", h.GetCameraLogs)
	mux.HandleFunc("POST /api/camera/logs", h.CreateCameraLogs)
	mux.HandleFunc("POST /api/3dadust", h.Create3DAdjustment)
	mux.HandleFunc("POST /api/camera/live", h.GetCameraLive) // single-shot fallback / testing
	mux.HandleFunc("GET /ws/camera/live", h.CameraLiveWS)    // real-time streaming
}
