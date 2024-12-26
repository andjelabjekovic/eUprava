package main

import (
	"context"
	"fmt"
	"healthcare-service/data"
	"healthcare-service/handlers"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	fmt.Println("Hello, World!")

	port := os.Getenv("HEALTHCARE_SERVICE_PORT")

	timeoutContext, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	logger := log.New(os.Stdout, "[res-api] ", log.LstdFlags)
	storeLogger := log.New(os.Stdout, "[res-store] ", log.LstdFlags)

	store, err := data.NewHealthCareRepo(timeoutContext, storeLogger)
	if err != nil {
		logger.Fatal(err)
	}
	defer store.DisconnectMongo(timeoutContext)
	store.Ping()

	healthCareHandler := handlers.NewHealthCareHandler(logger, store)

	// Inicijalizacija rutera i dodavanje middleware-a za sve zahteve
	router := mux.NewRouter()
	router.Use(MiddlewareContentTypeSet)

	getStudents := router.Methods(http.MethodGet).Subrouter()
	getStudents.HandleFunc("/students", healthCareHandler.GetAllUsers)

	insertStudent := router.Methods(http.MethodPost).Subrouter()
	insertStudent.HandleFunc("/students", healthCareHandler.InsertUser)
	insertStudent.Use(healthCareHandler.MiddlewareUserDeserialization)

	router.HandleFunc("/appointments", healthCareHandler.GetAllAppointments).Methods(http.MethodGet)
	router.HandleFunc("/therapies", healthCareHandler.GetAllTherapies).Methods(http.MethodGet)
	router.HandleFunc("/doneTherapies", healthCareHandler.GetDoneTherapiesFromFoodService).Methods(http.MethodGet)

	scheduleAppointment := router.Methods(http.MethodPost).Subrouter()
	scheduleAppointment.HandleFunc("/appointments/schedule", healthCareHandler.ScheduleAppointment)
	//scheduleAppointment.Use(healthCareHandler.MiddlewareAppointmentDeserialization)

	createAppointment := router.Methods(http.MethodPost).Subrouter()
	createAppointment.HandleFunc("/appointments", healthCareHandler.CreateAppointment)
	createAppointment.Use(healthCareHandler.MiddlewareAppointmentDeserialization)

	router.HandleFunc("/appointmentById", healthCareHandler.GetAppointmentByID).Methods(http.MethodGet)

	updateAppointment := router.Methods(http.MethodPatch).Subrouter()
	updateAppointment.HandleFunc("/appointment/update/{id}", healthCareHandler.UpdateAppointment)
	updateAppointment.Use(healthCareHandler.MiddlewareAppointmentDeserialization)

	router.HandleFunc("/appointment/delete", healthCareHandler.DeleteAppointment).Methods(http.MethodDelete)
	router.HandleFunc("/appointments/reserved", healthCareHandler.GetAllReservedAppointments).Methods(http.MethodGet)
	router.HandleFunc("/appointments/not_reserved", healthCareHandler.GetAllNotReservedAppointments).Methods(http.MethodGet)

	router.HandleFunc("/appointments/byUser", healthCareHandler.GetAllAppointmentsForUser).Methods(http.MethodGet)

	router.HandleFunc("/appointments/reservedByStudent", healthCareHandler.GetAllReservedAppointmentsForUser).Methods(http.MethodGet)
	
	saveTherapy := router.Methods(http.MethodPost).Subrouter()
	saveTherapy.HandleFunc("/therapy", healthCareHandler.SaveAndShareTherapyDataWithDietService)
	saveTherapy.Use(healthCareHandler.MiddlewareTherapyDeserialization)

	updateTherapy := router.Methods(http.MethodPut).Subrouter()
	updateTherapy.HandleFunc("/updateTherapy", healthCareHandler.UpdateTherapyFromFoodService)
	updateTherapy.Use(healthCareHandler.MiddlewareTherapyDeserialization)

	// Dodavanje ruta za terapije
	router.HandleFunc("/therapy/{id}", healthCareHandler.GetTherapyDataByID).Methods(http.MethodGet)
	router.HandleFunc("/therapy/{id}", healthCareHandler.DeleteTherapyData).Methods(http.MethodDelete)

	updateTherapy2 := router.Methods(http.MethodPut).Subrouter()
	updateTherapy2.HandleFunc("/therapy/{id}", healthCareHandler.UpdateTherapyData)
	updateTherapy2.Use(healthCareHandler.MiddlewareTherapyDeserialization)

	router.HandleFunc("/appointments/cancel", healthCareHandler.CancelAppointment).Methods("POST")

	router.HandleFunc("/student", healthCareHandler.GetUserByID).Methods(http.MethodGet)

	updateStudent := router.Methods(http.MethodPut).Subrouter()
	updateStudent.HandleFunc("/student/update/{id}", healthCareHandler.UpdateUser)
	updateStudent.Use(healthCareHandler.MiddlewareUserDeserialization)

	router.HandleFunc("/student/delete", healthCareHandler.DeleteUser).Methods(http.MethodDelete)

	updateHealthRecord := router.Methods(http.MethodPut).Subrouter()
	updateHealthRecord.HandleFunc("/healthrecords/{id}", healthCareHandler.UpdateHealthRecord)
	updateHealthRecord.Use(healthCareHandler.MiddlewareHealthRecordDeserialization)

	getHealthRecords := router.Methods(http.MethodGet).Subrouter()
	getHealthRecords.HandleFunc("/healthrecords", healthCareHandler.GetAllHealthRecords)

	router.HandleFunc("/healthrecords", healthCareHandler.GetHealthRecordByID).Methods(http.MethodGet)

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

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt)
	signal.Notify(sigCh, os.Kill)

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
