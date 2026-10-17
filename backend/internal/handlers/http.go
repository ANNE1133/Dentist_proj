package handlers

import (
	"net/http"

	"backend/internal/service"
)

const sessionCookieName = "dentist_session"

type Handler struct {
	service *service.Service
	auth    *service.AuthService
}

func NewHandler(svc *service.Service, auth *service.AuthService) *Handler {
	return &Handler{service: svc, auth: auth}
}

// RegisterRoutes wires HTTP method + path -> handler.
// NOTE: requires Go 1.22+ for method-prefixed mux patterns ("GET /path").
// Check your go.mod `go` directive - if it's below 1.22, tell me and I'll
// rewrite this using a method switch inside each handler instead.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("POST /api/login", h.Login)
	mux.HandleFunc("GET /api/session", h.Session)
	mux.HandleFunc("POST /api/logout", h.Logout)

	mux.Handle("GET /api/dentist", h.requireAuthentication(http.HandlerFunc(h.GetAllDentists)))
	mux.Handle("POST /api/dentist", h.requireAuthentication(http.HandlerFunc(h.CreateDentist)))
	mux.Handle("GET /api/dentist/{id}", h.requireAuthentication(http.HandlerFunc(h.GetDentist)))

	mux.Handle("GET /api/adjustitem", h.requireAuthentication(http.HandlerFunc(h.GetAllAdjustableItems)))
	mux.Handle("POST /api/adjustitem", h.requireAuthentication(http.HandlerFunc(h.CreateAdjustableItem)))
	mux.Handle("GET /api/adjustitem/{id}", h.requireAuthentication(http.HandlerFunc(h.GetAdjustableItembyID)))

	mux.Handle("GET /api/camera/logs", h.requireAuthentication(http.HandlerFunc(h.GetCameraLogs)))
	mux.Handle("POST /api/camera/logs", h.requireAuthentication(http.HandlerFunc(h.CreateCameraLogs)))
	mux.Handle("POST /api/3dadust", h.requireAuthentication(http.HandlerFunc(h.Create3DAdjustment)))
	mux.Handle("POST /api/camera/live", h.requireAuthentication(http.HandlerFunc(h.GetCameraLive))) // single-shot fallback / testing
	mux.Handle("GET /ws/camera/live", h.requireAuthentication(http.HandlerFunc(h.CameraLiveWS)))    // real-time streaming
}

func (h *Handler) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil || !h.auth.ValidateSession(cookie.Value) {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
