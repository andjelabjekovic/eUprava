package main

import (
	"context"
	"fmt"
	"food-service/data"
	"food-service/handlers"
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

	// Inicijalizacija loggera koji Ä‡e se koristiti, sa prefiksom i datumom za svaki log
	logger := log.New(os.Stdout, "[res-api] ", log.LstdFlags)
	storeLogger := log.New(os.Stdout, "[res-store] ", log.LstdFlags)

	// NoSQL: Inicijalizacija prodavnice proizvoda
	store, err := data.NewFoodServiceRepo(timeoutContext, storeLogger)
	if err != nil {
		logger.Fatal(err)
	}
	defer store.DisconnectMongo(timeoutContext)
	store.Ping()

	foodServiceHandler := handlers.NewFoodServiceHandler(logger, store)

	// Inicijalizacija rutera i dodavanje middleware-a za sve zahteve
	router := mux.NewRouter()
	router.Use(MiddlewareContentTypeSet)

	// Ruta za dobijanje liste hrane
	getFoodList := router.Methods(http.MethodGet).Subrouter()
	getFoodList.HandleFunc("/foods", foodServiceHandler.GetListFoodHandler)

	getFood := router.Methods(http.MethodGet).Subrouter()
	getFood.HandleFunc("/food/{id}", foodServiceHandler.GetFoodByIDHandler)

	// Kreiranje novog unosa hrane(radi)
	createFood := router.Methods(http.MethodPost).Subrouter()
	createFood.HandleFunc("/food", foodServiceHandler.CreateFoodHandler)
	createFood.Use(foodServiceHandler.MiddlewareFoodDeserialization)

	cancelOrder := router.Methods(http.MethodPut).Subrouter()
	cancelOrder.HandleFunc("/order/{id}/cancel", foodServiceHandler.CancelOrderHandler)

	getMyOrders := router.Methods(http.MethodGet).Subrouter()
	getMyOrders.HandleFunc("/my-orders", foodServiceHandler.GetMyOrdersHandler)

	//my orders
	//getMyOrders := router.Methods(http.MethodGet).Subrouter()
	//getMyOrders.HandleFunc("/order/my", foodServiceHandler.GetAllOrdersForUser)

	//getMyOrders := router.Methods(http.MethodGet).Subrouter()
	//getMyOrders.HandleFunc("/my-orders", foodServiceHandler.GetAllOrdersForUser)

	//router.HandleFunc("/my-orders", foodServiceHandler.GetAllOrdersForUser).Methods(http.MethodGet)

	getAllOrders := router.Methods(http.MethodGet).Subrouter()
	getAllOrders.HandleFunc("/order", foodServiceHandler.GetAllOrdersHandler)

	// Dohvatanje porudžbina ulogovanog korisnika sa statusO='Prihvacena' i statusO2='Neotkazana'
	//getMyOrders := router.Methods(http.MethodGet).Subrouter()
	//getMyOrders.HandleFunc("/my-orders", foodServiceHandler.GetAllMyOrdersHandler)

	// Dohvatanje prihvaćenih porudžbina
	getAcceptedOrders := router.Methods(http.MethodGet).Subrouter()
	getAcceptedOrders.HandleFunc("/accepted-orders", foodServiceHandler.GetAcceptedOrdersHandler)

	createOrder := router.Methods(http.MethodPost).Subrouter()
	createOrder.HandleFunc("/order", foodServiceHandler.CreateOrderHandler)
	createOrder.Use(foodServiceHandler.MiddlewareOrderDeserialization)

	updateOrderStatus := router.Methods(http.MethodPut).Subrouter()
	updateOrderStatus.HandleFunc("/order/{id}", foodServiceHandler.UpdateOrderStatusHandler)
	// Update postojeće hrane
	updateFood := router.Methods(http.MethodPut).Subrouter()
	updateFood.HandleFunc("/food/{id}", foodServiceHandler.UpdateFoodHandler)
	updateFood.Use(foodServiceHandler.MiddlewareFoodDeserialization)

	// Brisanje unosa hrane
	deleteFoodEntry := router.Methods(http.MethodDelete).Subrouter()
	deleteFoodEntry.HandleFunc("/food/{id}", foodServiceHandler.DeleteFoodHandler)

	// Ruta za dobijanje liste hrane(radi)
	getAllFood := router.Methods(http.MethodGet).Subrouter()
	getAllFood.HandleFunc("/food", foodServiceHandler.GetAllFood)

	getAllFoodForStudents := router.Methods(http.MethodGet).Subrouter()
	getAllFoodForStudents.HandleFunc("/studentsfood", foodServiceHandler.GetAllFoodOfStudents)
	//edit
	editFood := router.Methods(http.MethodPost).Subrouter()
	editFood.HandleFunc("/foods/{id}", foodServiceHandler.EditFood)
	editFood.Use(foodServiceHandler.MiddlewareFoodDeserialization)

	editFoodForStudent := router.Methods(http.MethodPost).Subrouter()
	editFoodForStudent.HandleFunc("/studentsfood", foodServiceHandler.EditFoodForStudent)
	editFoodForStudent.Use(foodServiceHandler.MiddlewareStudentDeserialization)

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

	// Inicijalizacija HTTP servera
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
		//s.logger.Println("Method [", h.Method, "] - Hit path :", h.URL.Path)

		rw.Header().Add("Content-Type", "application/json")
		rw.Header().Set("X-Content-Type-Options", "nosniff")
		rw.Header().Set("X-Frame-Options", "DENY")
		rw.Header().Set("Content-Security-Policy", "script-src 'self' https://code.jquery.com https://cdn.jsdelivr.net https://www.google.com https://www.gstatic.com 'unsafe-inline' 'unsafe-eval'; style-src 'self' https://code.jquery.com https://cdn.jsdelivr.net https://fonts.googleapis.com https://fonts.gstatic.com 'unsafe-inline'; font-src 'self' https://code.jquery.com https://cdn.jsdelivr.net https://fonts.googleapis.com https://fonts.gstatic.com; img-src 'self' data: https://code.jquery.com https://i.ibb.co;")

		next.ServeHTTP(rw, h)
	})
}
