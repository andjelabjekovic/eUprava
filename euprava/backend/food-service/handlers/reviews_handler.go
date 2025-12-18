package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"food-service/data"
	"food-service/middleware"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GET /food/{id}/reviews/summary
func (h *FoodServiceHandler) GetFoodReviewSummary(rw http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	foodID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(rw, "Invalid food id", http.StatusBadRequest)
		return
	}

	avg, rc, cc, err := h.foodServiceRepo.GetSummary(foodID)
	if err != nil {
		http.Error(rw, "Error reading summary", http.StatusInternalServerError)
		return
	}

	s := data.ReviewSummary{
		FoodID:       foodID,
		AvgRating:    avg,
		RatingCount:  rc,
		CommentCount: cc,
		CanReview:    false,
		MyRating:     0,
	}

	// Enrich if auth exists (AuthRequired middleware must be applied on this route in main if you want this to work)
	if uidStr, ok := middleware.GetUserID(r); ok {
		if userType, ok2 := middleware.GetUserType(r); ok2 && userType == "student" {
			userOID, err := primitive.ObjectIDFromHex(uidStr)
			if err == nil {
				can, err2 := h.foodServiceRepo.HasUserOrderedFood(userOID, foodID)
				if err2 == nil {
					s.CanReview = can
				}
				if my, exists, err3 := h.foodServiceRepo.GetMyRating(foodID, userOID); err3 == nil && exists {
					s.MyRating = my
				}
			}
		}
	}

	rw.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(rw).Encode(s)
}

// POST /food/{id}/reviews/rating  body: { "rating": 1..5 }
func (h *FoodServiceHandler) SetFoodRating(rw http.ResponseWriter, r *http.Request) {
	userType, _ := middleware.GetUserType(r)
	if userType != "student" {
		http.Error(rw, "Forbidden", http.StatusForbidden)
		return
	}

	uidStr, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(rw, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, err := primitive.ObjectIDFromHex(uidStr)
	if err != nil {
		http.Error(rw, "Invalid user id", http.StatusUnauthorized)
		return
	}

	idStr := mux.Vars(r)["id"]
	foodID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(rw, "Invalid food id", http.StatusBadRequest)
		return
	}

	can, err := h.foodServiceRepo.HasUserOrderedFood(userID, foodID)
	if err != nil {
		http.Error(rw, "Error checking orders", http.StatusInternalServerError)
		return
	}
	if !can {
		http.Error(rw, "Forbidden", http.StatusForbidden)
		return
	}

	var req data.SetRatingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(rw, "Bad request", http.StatusBadRequest)
		return
	}

	err = h.foodServiceRepo.UpsertRating(foodID, userID, req.Rating)
	if err != nil {
		http.Error(rw, "Cannot save rating", http.StatusBadRequest)
		return
	}

	avg, rc, cc, err := h.foodServiceRepo.GetSummary(foodID)
	if err != nil {
		http.Error(rw, "Saved but cannot read summary", http.StatusInternalServerError)
		return
	}

	out := data.ReviewSummary{
		FoodID:       foodID,
		AvgRating:    avg,
		RatingCount:  rc,
		CommentCount: cc,
		CanReview:    true,
		MyRating:     req.Rating,
	}

	rw.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(rw).Encode(out)
}

// GET /food/{id}/reviews/comments?limit=50
func (h *FoodServiceHandler) ListFoodComments(rw http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	foodID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(rw, "Invalid food id", http.StatusBadRequest)
		return
	}

	comments, err := h.foodServiceRepo.ListComments(foodID, 50)
	if err != nil {
		http.Error(rw, "Error reading comments", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(rw).Encode(comments)
}

// POST /food/{id}/reviews/comments body: { "text": "..." }
func (h *FoodServiceHandler) AddFoodComment(rw http.ResponseWriter, r *http.Request) {
	userType, _ := middleware.GetUserType(r)
	if userType != "student" {
		http.Error(rw, "Forbidden", http.StatusForbidden)
		return
	}

	uidStr, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(rw, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, err := primitive.ObjectIDFromHex(uidStr)
	if err != nil {
		http.Error(rw, "Invalid user id", http.StatusUnauthorized)
		return
	}

	idStr := mux.Vars(r)["id"]
	foodID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(rw, "Invalid food id", http.StatusBadRequest)
		return
	}

	can, err := h.foodServiceRepo.HasUserOrderedFood(userID, foodID)
	if err != nil {
		http.Error(rw, "Error checking orders", http.StatusInternalServerError)
		return
	}
	if !can {
		http.Error(rw, "Forbidden", http.StatusForbidden)
		return
	}

	var req data.AddCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(rw, "Bad request", http.StatusBadRequest)
		return
	}

	text := strings.TrimSpace(req.Text)
	if text == "" {
		http.Error(rw, "Empty comment", http.StatusBadRequest)
		return
	}

	author := middleware.GetFullName(r)

	if err := h.foodServiceRepo.AddComment(foodID, userID, author, text); err != nil {
		http.Error(rw, "Cannot save comment", http.StatusBadRequest)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(rw).Encode(map[string]any{"ok": true})
}

// POST /foods/reviews/summaries  body: { "foodIds": ["hex","hex"] }
func (h *FoodServiceHandler) BatchFoodSummaries(rw http.ResponseWriter, r *http.Request) {
	var req data.BatchSummaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(rw, "Bad request", http.StatusBadRequest)
		return
	}

	ids := make([]primitive.ObjectID, 0, len(req.FoodIds))
	for _, s := range req.FoodIds {
		oid, err := primitive.ObjectIDFromHex(s)
		if err != nil {
			continue
		}
		ids = append(ids, oid)
	}

	m, err := h.foodServiceRepo.GetBatchSummaries(ids)
	if err != nil {
		http.Error(rw, "Error reading summaries", http.StatusInternalServerError)
		return
	}

	out := map[string]data.ReviewSummary{}
	for k, v := range m {
		out[k.Hex()] = v
	}

	rw.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(rw).Encode(out)
}
