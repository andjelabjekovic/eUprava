package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	"university--service/client"
	"university--service/models"
	"university--service/store"
)

type CateringHandler struct {
	Logger *log.Logger
	Store  *store.CateringStore
	Food   *client.FoodClient
}

func NewCateringHandler(l *log.Logger, s *store.CateringStore, fc *client.FoodClient) *CateringHandler {
	return &CateringHandler{
		Logger: l,
		Store:  s,
		Food:   fc,
	}
}

// POST /university/catering
func (h *CateringHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateCateringRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	requestId := req.RequestID
	if requestId == "" {
		requestId = fmt.Sprintf("cat-%d", time.Now().UnixNano())
	}

	now := time.Now()
	cr := &models.CateringRequest{
		RequestID: requestId,
		Foods:     req.Foods,
		Note:      req.Note,
		Status:    "NEMA",
		CreatedAt: now,
		UpdatedAt: now,
	}

	h.Store.Save(cr)

	// Send to FoodService
	if err := h.Food.SendCateringNotification(cr); err != nil {
		h.Logger.Println("[UNIVERSITY] Food service unreachable:", err)
		http.Error(w, "Food service unreachable", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(cr)
}

// GET /university/catering/{id}
func (h *CateringHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	// /university/catering/{id}
	if len(parts) != 4 {
		http.Error(w, "Invalid URL format. Expected /university/catering/{id}", http.StatusBadRequest)
		return
	}

	id := parts[3]
	cr, ok := h.Store.Get(id)
	if !ok {
		http.Error(w, "Catering request not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(cr)
}

// GET /university/catering
func (h *CateringHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	list := h.Store.List()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(list)
}

// PUT /university/catering/{id}/status
func (h *CateringHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	// /university/catering/{id}/status
	if len(parts) != 5 || parts[4] != "status" {
		http.Error(w, "Invalid URL format. Expected /university/catering/{id}/status", http.StatusBadRequest)
		return
	}

	id := parts[3]

	var req models.UpdateCateringStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	if req.Status != "IMA" && req.Status != "NEMA" {
		http.Error(w, "Invalid status (IMA | NEMA)", http.StatusBadRequest)
		return
	}

	cr, err := h.Store.UpdateStatus(id, req.Status)
	if err != nil {
		http.Error(w, "Catering request not found", http.StatusNotFound)
		return
	}
	cr.UpdatedAt = time.Now()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"message":   "Catering status updated",
		"requestId": id,
		"status":    cr.Status,
	})
}
