package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"university--service/client"
	"university--service/handlers"
	"university--service/store"
)

type UpdateStudentStatusRequest struct {
	Status string `json:"status"` // "BUDZET" ili "SAMOFINANSIRANJE"
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8006"
	}

	foodURL := os.Getenv("FOOD_SERVICE_URL")
	if foodURL == "" {
		foodURL = "http://food_service:8003" // u docker mreži
		// lokalno: "http://localhost:8003"
	}

	logger := log.New(os.Stdout, "[UNIVERSITY] ", log.LstdFlags)

	// ✅ deps za catering
	cateringStore := store.NewCateringStore()
	foodClient := client.NewFoodClient(foodURL)
	cateringHandler := handlers.NewCateringHandler(logger, cateringStore, foodClient)

	mux := http.NewServeMux()

	// health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("university--service OK"))
	})

	// =========================
	// ✅ STUDENT STATUS (TVOJ KOD 1/1)
	// PUT /university/student/{id}/status
	// =========================
	mux.HandleFunc("/university/student/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// očekivani path: /university/student/{id}/status
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) != 5 || parts[4] != "status" {
			http.Error(w, "Invalid URL format. Expected /university/student/{id}/status", http.StatusBadRequest)
			return
		}

		userID := parts[3]
		if userID == "" {
			http.Error(w, "Missing user ID", http.StatusBadRequest)
			return
		}

		var req UpdateStudentStatusRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid body", http.StatusBadRequest)
			return
		}
		if req.Status != "BUDZET" && req.Status != "SAMOFINANSIRANJE" {
			http.Error(w, "Invalid status (BUDZET | SAMOFINANSIRANJE)", http.StatusBadRequest)
			return
		}

		log.Println("[UNIVERSITY] changing status for user:", userID, "=>", req.Status)

		// poziv Food servisa: PUT /student/{id}/status
		payload, _ := json.Marshal(req)
		clientHTTP := &http.Client{Timeout: 10 * time.Second}

		foodEndpoint := foodURL + "/student/" + userID + "/status"
		foodReq, err := http.NewRequest(http.MethodPut, foodEndpoint, bytes.NewBuffer(payload))
		if err != nil {
			http.Error(w, "Cannot create request", http.StatusInternalServerError)
			return
		}
		foodReq.Header.Set("Content-Type", "application/json")

		resp, err := clientHTTP.Do(foodReq)
		if err != nil {
			log.Println("Error calling food service:", err)
			http.Error(w, "Food service unreachable", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			body, _ := io.ReadAll(resp.Body)
			log.Println("Food service returned:", resp.StatusCode, string(body))
			http.Error(w, "Food service returned error", http.StatusBadGateway)
			return
		}

		// ✅ prosledi Food response nazad Postmanu
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.Copy(w, resp.Body)
	})

	// =========================
	// ✅ CATERING (POSTMAN → University → FoodService notification)
	// =========================

	// POST /university/catering
	//mux.HandleFunc("/university/catering", cateringHandler.Create)
	mux.HandleFunc("/university/catering", func(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		cateringHandler.Create(w, r)
		return
	}
	if r.Method == http.MethodGet {
		cateringHandler.ListAll(w, r)
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
})

	// GET /university/catering/{id}
	// PUT /university/catering/{id}/status
	mux.HandleFunc("/university/catering/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			cateringHandler.GetByID(w, r)
			return
		}
		if r.Method == http.MethodPut {
			cateringHandler.UpdateStatus(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	logger.Printf("university--service started on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		logger.Fatal(err)
	}
}
