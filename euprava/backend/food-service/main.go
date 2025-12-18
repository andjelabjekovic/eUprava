package main

import (
	"context"
	"fmt"
	"food-service/data"
	"food-service/handlers"
	"food-service/middleware"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"syscall"

	"github.com/gorilla/mux"
)

func main() {
	fmt.Println("Hello, World!")

	port := os.Getenv("FOOD_SERVICE_PORT")

	timeoutContext, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	// Inicijalizacija loggera
	logger := log.New(os.Stdout, "[res-api] ", log.LstdFlags)
	storeLogger := log.New(os.Stdout, "[res-store] ", log.LstdFlags)

	// Repo
	store, err := data.NewFoodServiceRepo(timeoutContext, storeLogger)
	if err != nil {
		logger.Fatal(err)
	}
	defer store.DisconnectMongo(timeoutContext)
	store.Ping()

	// ✅ REVIEWS: indeksi
	if err := store.EnsureReviewIndexes(); err != nil {
		logger.Println("Warning: cannot ensure review indexes:", err)
	}

	foodServiceHandler := handlers.NewFoodServiceHandler(logger, store)

	// Router + middleware
	router := mux.NewRouter()
	router.Use(MiddlewareContentTypeSet)
uploadDir := os.Getenv("UPLOAD_DIR")
if uploadDir == "" {
    uploadDir = "./uploads"
}

if err := os.MkdirAll(uploadDir, 0755); err != nil {
    logger.Fatal("Cannot create upload dir:", err)
}

router.PathPrefix("/uploads/").
  Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("/uploads/"))))



	uploadFoodImage := router.Methods(http.MethodPost).Subrouter()
	uploadFoodImage.HandleFunc("/food/{id}/image", foodServiceHandler.UploadFoodImageHandler)

	// Foods
	getFoodList := router.Methods(http.MethodGet).Subrouter()
	getFoodList.HandleFunc("/foods", foodServiceHandler.GetListFoodHandler)

	getFood := router.Methods(http.MethodGet).Subrouter()
	getFood.HandleFunc("/food/{id}", foodServiceHandler.GetFoodByIDHandler)

	createFood := router.Methods(http.MethodPost).Subrouter()
	createFood.HandleFunc("/food", foodServiceHandler.CreateFoodHandler)
	createFood.Use(foodServiceHandler.MiddlewareFoodDeserialization)

	updateFood := router.Methods(http.MethodPut).Subrouter()
	updateFood.HandleFunc("/food/{id}", foodServiceHandler.UpdateFoodHandler)
	updateFood.Use(foodServiceHandler.MiddlewareFoodDeserialization)

	deleteFoodEntry := router.Methods(http.MethodDelete).Subrouter()
	deleteFoodEntry.HandleFunc("/food/{id}", foodServiceHandler.DeleteFoodHandler)

	getAllFood := router.Methods(http.MethodGet).Subrouter()
	getAllFood.HandleFunc("/food", foodServiceHandler.GetAllFood)

	// Orders
	cancelOrder := router.Methods(http.MethodPut).Subrouter()
	cancelOrder.HandleFunc("/order/{id}/cancel", foodServiceHandler.CancelOrderHandler)

	getMyOrders := router.Methods(http.MethodGet).Subrouter()
	getMyOrders.HandleFunc("/my-orders", foodServiceHandler.GetMyOrdersHandler)

	getAllOrders := router.Methods(http.MethodGet).Subrouter()
	getAllOrders.HandleFunc("/order", foodServiceHandler.GetAllOrdersHandler)

	getAcceptedOrders := router.Methods(http.MethodGet).Subrouter()
	getAcceptedOrders.HandleFunc("/accepted-orders", foodServiceHandler.GetAcceptedOrdersHandler)

	createOrder := router.Methods(http.MethodPost).Subrouter()
	createOrder.HandleFunc("/order", foodServiceHandler.CreateOrderHandler)
	createOrder.Use(foodServiceHandler.MiddlewareOrderDeserialization)

	updateOrderStatus := router.Methods(http.MethodPut).Subrouter()
	updateOrderStatus.HandleFunc("/order/{id}", foodServiceHandler.UpdateOrderStatusHandler)

	// Students food / edit legacy
	getAllFoodForStudents := router.Methods(http.MethodGet).Subrouter()
	getAllFoodForStudents.HandleFunc("/studentsfood", foodServiceHandler.GetAllFoodOfStudents)

	editFood := router.Methods(http.MethodPost).Subrouter()
	editFood.HandleFunc("/foods/{id}", foodServiceHandler.EditFood)
	editFood.Use(foodServiceHandler.MiddlewareFoodDeserialization)

	editFoodForStudent := router.Methods(http.MethodPost).Subrouter()
	editFoodForStudent.HandleFunc("/studentsfood", foodServiceHandler.EditFoodForStudent)
	editFoodForStudent.Use(foodServiceHandler.MiddlewareStudentDeserialization)

	// Therapies
	getTherapies := router.Methods(http.MethodGet).Subrouter()
	getTherapies.HandleFunc("/therapies", foodServiceHandler.GetTherapies)

	editTherapy := router.Methods(http.MethodPut).Subrouter()
	editTherapy.HandleFunc("/therapy/{therapyId}/approve", foodServiceHandler.ApproveTherapy)
	editTherapy.Use(foodServiceHandler.MiddlewareFoodDeserialization)

	saveTherapy := router.Methods(http.MethodPost).Subrouter()
	saveTherapy.HandleFunc("/therapy", foodServiceHandler.SaveTherapy)
	saveTherapy.Use(foodServiceHandler.MiddlewareTherapyDeserialization)

	clearAllTherapy := router.Methods(http.MethodDelete).Subrouter()
	clearAllTherapy.HandleFunc("/therapy", foodServiceHandler.ClearTherapiesList)

	updateTherapyStatus := router.Methods(http.MethodPut).Subrouter()
	updateTherapyStatus.HandleFunc("/therapy/{id}", foodServiceHandler.UpdateTherapyStatus)

	// =========================
	// ✅ REVIEWS ROUTES (ALL IN MAIN)
	// =========================

	// GET summary (optional auth if you want CanReview/MyRating enriched)
	getReviewSummary := router.Methods(http.MethodGet).Subrouter()
	getReviewSummary.Use(middleware.AuthRequired) // <- skini ovu liniju ako hoćeš da radi i bez tokena
	getReviewSummary.HandleFunc("/food/{id}/reviews/summary", foodServiceHandler.GetFoodReviewSummary)

	// POST rating (must be auth + student)
	setRating := router.Methods(http.MethodPost).Subrouter()
	setRating.Use(middleware.AuthRequired)
	setRating.HandleFunc("/food/{id}/reviews/rating", foodServiceHandler.SetFoodRating)

	// GET comments (public)
	listComments := router.Methods(http.MethodGet).Subrouter()
	listComments.HandleFunc("/food/{id}/reviews/comments", foodServiceHandler.ListFoodComments)

	// POST comment (must be auth + student)
	addComment := router.Methods(http.MethodPost).Subrouter()
	addComment.Use(middleware.AuthRequired)
	addComment.HandleFunc("/food/{id}/reviews/comments", foodServiceHandler.AddFoodComment)

	// Batch summaries for food list
	batchSummaries := router.Methods(http.MethodPost).Subrouter()
	batchSummaries.HandleFunc("/foods/reviews/summaries", foodServiceHandler.BatchFoodSummaries)

	// Server
	server := http.Server{
		Addr:         ":" + port,
		Handler:      router,
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Println("Server listening on port", port)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			logger.Fatal(err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	sig := <-sigCh
	logger.Println("Received terminate, graceful shutdown", sig)

	if err := server.Shutdown(timeoutContext); err != nil {
		logger.Fatal("Cannot gracefully shutdown...", err)
	}
	logger.Println("Server stopped")
}

func MiddlewareContentTypeSet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		rw.Header().Set("X-Content-Type-Options", "nosniff")
		rw.Header().Set("X-Frame-Options", "DENY")
		rw.Header().Set("Content-Security-Policy",
			"script-src 'self' https://code.jquery.com https://cdn.jsdelivr.net https://www.google.com https://www.gstatic.com 'unsafe-inline' 'unsafe-eval'; "+
				"style-src 'self' https://code.jquery.com https://cdn.jsdelivr.net https://fonts.googleapis.com https://fonts.gstatic.com 'unsafe-inline'; "+
				"font-src 'self' https://code.jquery.com https://cdn.jsdelivr.net https://fonts.googleapis.com https://fonts.gstatic.com; "+
				"img-src 'self' data: https://code.jquery.com https://i.ibb.co;")

		next.ServeHTTP(rw, h)
	})
}
