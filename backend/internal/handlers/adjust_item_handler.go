package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"backend/internal/domain"

)

// GetAllAdjustableItems handles POST /api/adjustitem to create an adjust item and GET /api/adjustitem to list all adjust items.
func (h *Handler) GetAllAdjustableItems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.Method == http.MethodGet {
		items, err := h.service.ListAdjustItems()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(items)
	}
}

// GetAdjustableItembyID handles GET /api/adjustitem/:id to fetch one adjust item by id.
func (h *Handler) GetAdjustableItembyID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.Error(w, "adjust item id is required", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		http.Error(w, "invalid adjust item id", http.StatusBadRequest)
		return
	}
	item, err := h.service.GetAdjustItemByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(item)
}


// GetAllAdjustableItems handles POST /api/adjustitem to create an adjust item and GET /api/adjustitem to list all adjust items.
func (h *Handler) CreateAdjustableItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
		var payload domain.AdjustableItem
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		created, err := h.service.CreateAdjustItem(payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(created)
}