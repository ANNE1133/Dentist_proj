package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"backend/internal/domain"
)

type CreateDentistResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	DentistID int    `json:"dentistId"`
}

func (h *Handler) CreateDentist(w http.ResponseWriter, r *http.Request) {

	// จำกัด request size เช่น 10 MB
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "invalid multipart form", http.StatusBadRequest)
		return
	}

	dentistName := strings.TrimSpace(r.FormValue("dentistName"))
	if dentistName == "" {
		http.Error(w, "dentistName is required", http.StatusBadRequest)
		return
	}

	// รับ image
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "image is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// ตรวจ MIME type
	contentType := header.Header.Get("Content-Type")
	if contentType != "image/jpeg" &&
		contentType != "image/png" &&
		contentType != "image/webp" {
		http.Error(w, "unsupported image type", http.StatusBadRequest)
		return
	}

	payload := domain.Dentist{
		DentistName: dentistName,
	}

	result, err := h.service.RegisterDentistWithImage(
		r.Context(),
		payload,
		file,
		header.Filename,
		contentType,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(CreateDentistResponse{
		Success:   result.Success,
		Message:   "dentist registered successfully",
		DentistID: result.DentistID,
	})
}

// GetAllDentists handles GET /api/dentist to list all dentist records.
func (h *Handler) GetAllDentists(w http.ResponseWriter, r *http.Request) {
	dentists, err := h.service.ListDentistRecords()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dentists)
}

// GetDentist handles GET /api/dentist/:id to fetch a single dentist record by id.
func (h *Handler) GetDentist(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.Error(w, "dentist id is required", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		http.Error(w, "invalid dentist id", http.StatusBadRequest)
		return
	}
	dentist, err := h.service.GetDentistRecordByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dentist)
}