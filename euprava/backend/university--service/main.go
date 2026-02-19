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
		foodURL = "http://food-service:8000" // u docker mreži
		// lokalno: "http://localhost:8000" ili host port koji mapira na 8000
	}

	mux := http.NewServeMux()

	// health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("university--service OK"))
	})

	// PUT /university/student/{id}/status
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
		client := &http.Client{Timeout: 10 * time.Second}

		foodEndpoint := foodURL + "/student/" + userID + "/status"
		foodReq, err := http.NewRequest(http.MethodPut, foodEndpoint, bytes.NewBuffer(payload))
		if err != nil {
			http.Error(w, "Cannot create request", http.StatusInternalServerError)
			return
		}
		foodReq.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(foodReq)
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

	log.Printf("university--service started on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
