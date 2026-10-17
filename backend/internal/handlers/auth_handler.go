package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"backend/internal/service"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var credentials loginRequest
	if err := decoder.Decode(&credentials); err != nil {
		http.Error(w, "invalid login request", http.StatusBadRequest)
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		http.Error(w, "invalid login request", http.StatusBadRequest)
		return
	}
	if credentials.Username == "" || credentials.Password == "" {
		http.Error(w, "username and password are required", http.StatusBadRequest)
		return
	}

	token, err := h.auth.Login(credentials.Username, credentials.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAuthNotConfigured):
			http.Error(w, "login is not configured", http.StatusServiceUnavailable)
		case errors.Is(err, service.ErrInvalidCredentials):
			http.Error(w, "invalid username or password", http.StatusUnauthorized)
		default:
			http.Error(w, "unable to create login session", http.StatusInternalServerError)
		}
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(service.SessionDuration.Seconds()),
	})
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (h *Handler) Session(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || !h.auth.ValidateSession(cookie.Value) {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"authenticated": true})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		h.auth.Logout(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
	w.WriteHeader(http.StatusNoContent)
}
