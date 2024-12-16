package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"food-service/data"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FoodServiceHandler struct {
	logger          *log.Logger
	foodServiceRepo *data.FoodServiceRepo
}

type KeyProduct struct{}
type KeyFood struct{}
type KeyOrder struct{}

func NewFoodServiceHandler(l *log.Logger, r *data.FoodServiceRepo) *FoodServiceHandler {
	return &FoodServiceHandler{l, r}
}

// GetListFoodHandler vraća sve unose hrane iz baze
func (h *FoodServiceHandler) GetListFoodHandler(rw http.ResponseWriter, r *http.Request) {
	foodList, err := h.foodServiceRepo.GetListFood()
	if err != nil {
		h.logger.Print("Database exception: ", err)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("Error retrieving food entries."))
		return
	}

	// Konvertuj listu hrane u JSON i pošalji klijentu
	rw.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(rw).Encode(foodList)
	if err != nil {
		h.logger.Print("Error converting food list to JSON: ", err)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("Error converting food list to JSON."))
		return
	}
}
func (h *FoodServiceHandler) CreateFoodHandler(rw http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("cookId")
	if userID == "" {
		http.Error(rw, "cookId is required", http.StatusBadRequest)
		return
	}
	// Konverzija userID (cookId) u ObjectID
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		h.logger.Println("Invalid cookId:", err)
		http.Error(rw, "Invalid cookId", http.StatusBadRequest)
		return
	}
	
	foodData := r.Context().Value(KeyFood{}).(*data.Food)
	fmt.Printf("Received food entry: %+v\n", foodData)

	foodData.UserID = oid

	err = h.foodServiceRepo.CreateFoodEntry(r, foodData)
	if err != nil {
		h.logger.Print("Database exception: ", err)
		http.Error(rw, "Error creating food entry.", http.StatusInternalServerError)
		return
	}
	// Postavi status kod 201 Created
	rw.WriteHeader(http.StatusCreated)
}

func (h *FoodServiceHandler) GetFoodByIDHandler(rw http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    idStr, ok := vars["id"]
    if !ok {
        http.Error(rw, "ID is required", http.StatusBadRequest)
        return
    }

    // Konverzija idStr u ObjectID
    oid, err := primitive.ObjectIDFromHex(idStr)
    if err != nil {
        h.logger.Println("Invalid ID:", err)
        http.Error(rw, "Invalid ID", http.StatusBadRequest)
        return
    }

    food, err := h.foodServiceRepo.GetFoodByID(r, oid)
    if err != nil {
        h.logger.Print("Database exception: ", err)
        http.Error(rw, "Error fetching food entry.", http.StatusInternalServerError)
        return
    }

    if food == nil {
        // Nema dokumenata za taj ID
        http.Error(rw, "Food not found", http.StatusNotFound)
        return
    }

    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(http.StatusOK)
    json.NewEncoder(rw).Encode(food)
}

// CreateFoodHandler kreira novi unos hrane sa stanjem postavljenim na 'Neporucena'
/*
func (h *FoodServiceHandler) OrderHandler(rw http.ResponseWriter, r *http.Request) {
	// Preuzmi podatke o hrani iz konteksta
	orderData := r.Context().Value(KeyOrder{}).(*data.Order)

	fmt.Printf("Received food entry: %+v\n", orderData)

	// Kreiraj novi unos hrane koristeći metodu iz repo
	err := h.foodServiceRepo.CreateOrder(r, orderData)
	if err != nil {
		h.logger.Print("Database exception: ", err)
		http.Error(rw, "Error creating food entry.", http.StatusInternalServerError)
		return
	}

	// Postavi status kod 201 Created
	rw.WriteHeader(http.StatusCreated)

	// Definiši strukturu odgovora sa porukom i podacima o kreiranoj hrani
	response := map[string]interface{}{
		"message":  "Food entry created successfully",
		"foodName": orderData.Food.FoodName,
	}

	// Pošalji JSON odgovor nazad klijentu
	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(response)
}*/

func (h *FoodServiceHandler) MiddlewareOrderDeserialization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		orderData := &data.Order{}

		// Deserijalizuj JSON u orderData
		err := json.NewDecoder(r.Body).Decode(orderData)
		if err != nil {
			h.logger.Println("Unable to decode JSON:", err)
			http.Error(rw, "Unable to decode JSON", http.StatusBadRequest)
			return
		}

		// Postavi orderData u kontekst
		ctx := context.WithValue(r.Context(), KeyOrder{}, orderData)
		r = r.WithContext(ctx)

		next.ServeHTTP(rw, r)
	})
}

// UpdateOrderStatusHandler ažurira status porudžbine na 'Prihvacena'
func (h *FoodServiceHandler) UpdateOrderStatusHandler(rw http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    idStr, ok := vars["id"]
    if !ok {
        http.Error(rw, "ID is required", http.StatusBadRequest)
        return
    }

    // Konverzija idStr u ObjectID
    orderID, err := primitive.ObjectIDFromHex(idStr)
    if err != nil {
        h.logger.Println("Invalid ID:", err)
        http.Error(rw, "Invalid ID", http.StatusBadRequest)
        return
    }

    // Ažuriraj status porudžbine na 'Prihvacena'
    err = h.foodServiceRepo.UpdateOrderStatus(orderID, data.Prihvacena)
    if err != nil {
        if err.Error() == "order not found" {
            http.Error(rw, "Order not found", http.StatusNotFound)
            return
        }
        h.logger.Print("Database exception:", err)
        http.Error(rw, "Error updating order status.", http.StatusInternalServerError)
        return
    }

    // Ako je sve prošlo dobro, vrati odgovor
    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(http.StatusOK)
    response := map[string]interface{}{
        "message": "Order status updated successfully",
        "orderId": orderID.Hex(),
        "status":  data.Prihvacena,
    }
    json.NewEncoder(rw).Encode(response)
}


// GetAcceptedOrdersHandler vraća sve porudžbine čiji je statusO='Prihvacena'
func (h *FoodServiceHandler) GetAcceptedOrdersHandler(rw http.ResponseWriter, r *http.Request) {
    acceptedOrders, err := h.foodServiceRepo.GetAcceptedOrders()
    if err != nil {
        h.logger.Println("Error retrieving accepted orders:", err)
        http.Error(rw, "Error retrieving accepted orders.", http.StatusInternalServerError)
        return
    }

    if len(acceptedOrders) == 0 {
        http.Error(rw, "No accepted orders found.", http.StatusNotFound)
        return
    }

    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(http.StatusOK)
    if err := json.NewEncoder(rw).Encode(acceptedOrders); err != nil {
        h.logger.Println("Error encoding accepted orders to JSON:", err)
        http.Error(rw, "Error encoding data to JSON.", http.StatusInternalServerError)
        return
    }
}


func (h *FoodServiceHandler) CancelOrderHandler(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderIDStr := vars["id"]

	orderID, err := primitive.ObjectIDFromHex(orderIDStr)
	if err != nil {
		h.logger.Printf("Invalid order ID: %v", err)
		http.Error(rw, "Invalid order ID format", http.StatusBadRequest)
		return
	}

	err = h.foodServiceRepo.CancelOrder(orderID)
	if err != nil {
		if err.Error() == "order not found" {
			h.logger.Printf("Order not found: %v", orderIDStr)
			http.Error(rw, "Order not found", http.StatusNotFound)
			return
		}
		h.logger.Printf("Error updating order status: %v", err)
		http.Error(rw, "Error updating order status", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(`{"message": "Order canceled successfully"}`))
}



func (h *FoodServiceHandler) GetAllOrdersHandler(rw http.ResponseWriter, r *http.Request) {
    orders, err := h.foodServiceRepo.GetAllOrders()
    if err != nil {
        h.logger.Println("Error fetching orders:", err)
        http.Error(rw, "Error fetching orders", http.StatusInternalServerError)
        return
    }

    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(http.StatusOK)
    json.NewEncoder(rw).Encode(orders)
}

func (h *FoodServiceHandler) CreateOrderHandler(rw http.ResponseWriter, r *http.Request) {
	// Izvuci userId iz query stringa (analogno cookId za hranu)
	userID := r.URL.Query().Get("userId")
	if userID == "" {
		http.Error(rw, "userId is required", http.StatusBadRequest)
		return
	}

	// Konvertuj userID u ObjectID
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		h.logger.Println("Invalid userId:", err)
		http.Error(rw, "Invalid userId", http.StatusBadRequest)
		return
	}

	// Izvuci orderData iz konteksta
	orderData, ok := r.Context().Value(KeyOrder{}).(*data.Order)
	if !ok || orderData == nil {
		h.logger.Println("Order data not found in context")
		http.Error(rw, "Order data not found in context", http.StatusInternalServerError)
		return
	}

	// Postavi UserID
	orderData.UserID = oid

	// Kreiraj porudžbinu kroz repo sloj
	err = h.foodServiceRepo.CreateOrderEntry(r, orderData)
	if err != nil {
		h.logger.Print("Database exception: ", err)
		http.Error(rw, "Error creating order.", http.StatusInternalServerError)
		return
	}

	// Postavi status kod 201 Created
	rw.WriteHeader(http.StatusCreated)
}
// GetAllMyOrdersHandler vraća porudžbine ulogovanog korisnika sa statusO='Prihvacena' i statusO2='Neotkazana'
/*func (h *FoodServiceHandler) GetAllMyOrdersHandler(rw http.ResponseWriter, r *http.Request) {
    // Dobavi ulogovanog korisnika
    user, err := h.foodServiceRepo.GetLoggedUser(r)
    if err != nil {
        h.logger.Println("Error retrieving logged user:", err)
        http.Error(rw, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Dobavi porudžbine za korisnika sa zadatim uslovima
    orders, err := h.foodServiceRepo.GetAllMyOrders(user.ID)
    if err != nil {
        h.logger.Println("Error fetching user's orders:", err)
        http.Error(rw, "Error fetching orders", http.StatusInternalServerError)
        return
    }

    // Vrati porudžbine kao JSON odgovor
    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(http.StatusOK)
    json.NewEncoder(rw).Encode(orders)
}
*/

/*func (h *FoodServiceHandler) GetAllOrdersForUser(rw http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(rw, "User ID is required", http.StatusBadRequest)
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		h.logger.Print("Invalid user ID format: ", err)
		http.Error(rw, "Invalid user ID", http.StatusBadRequest)
		return
	}

	orders, err := h.foodServiceRepo.GetAllOrdersForUser(userID)
	if err != nil {
		h.logger.Print("Database exception: ", err)
		http.Error(rw, "Error retrieving orders.", http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		http.Error(rw, "No orders found for the user.", http.StatusNotFound)
		return
	}

	err = json.NewEncoder(rw).Encode(orders)
	if err != nil {
		h.logger.Print("Error encoding orders to JSON: ", err)
		http.Error(rw, "Error encoding orders to JSON.", http.StatusInternalServerError)
		return
	}
}*/
func (h *FoodServiceHandler) GetMyOrdersHandler(rw http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(rw, "User ID is required", http.StatusBadRequest)
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		h.logger.Println("Invalid user ID format:", err)
		http.Error(rw, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	myOrders, err := h.foodServiceRepo.GetMyOrders(userID)
	if err != nil {
		h.logger.Println("Error retrieving orders for user:", err)
		http.Error(rw, "Error retrieving orders for user.", http.StatusInternalServerError)
		return
	}

	if len(myOrders) == 0 {
		http.Error(rw, "No orders found for user.", http.StatusNotFound)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(rw).Encode(myOrders); err != nil {
		h.logger.Println("Error encoding orders to JSON:", err)
		http.Error(rw, "Error encoding data to JSON.", http.StatusInternalServerError)
		return
	}
}

func (h *FoodServiceHandler) UpdateFoodHandler(rw http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    idStr, ok := vars["id"]
    if !ok {
        http.Error(rw, "ID is required", http.StatusBadRequest)
        return
    }

    // Konverzija idStr u ObjectID
    foodID, err := primitive.ObjectIDFromHex(idStr)
    if err != nil {
        h.logger.Println("Invalid ID:", err)
        http.Error(rw, "Invalid ID", http.StatusBadRequest)
        return
    }

    // Preuzmi podatke o hrani iz konteksta (nova vrednost foodName)
    foodData := r.Context().Value(KeyFood{}).(*data.Food)
    if foodData.FoodName == "" {
        http.Error(rw, "foodName cannot be empty", http.StatusBadRequest)
        return
    }

    // Pozovi repo metodu za ažuriranje unosa
    err = h.foodServiceRepo.UpdateFoodEntry(r, foodID, foodData)
    if err != nil {
        h.logger.Print("Database exception: ", err)
        http.Error(rw, "Error updating food entry.", http.StatusInternalServerError)
        return
    }

    // Ako je sve prošlo dobro
    rw.WriteHeader(http.StatusOK)
    response := map[string]interface{}{
        "message":  "Food entry updated successfully",
        "foodId":   foodID.Hex(),
        "foodName": foodData.FoodName,
    }
    rw.Header().Set("Content-Type", "application/json")
    json.NewEncoder(rw).Encode(response)
}


// mongo
// editFood
func (r *FoodServiceHandler) EditFoodForStudent(rw http.ResponseWriter, h *http.Request) {
	// Parse request parameters
	vars := mux.Vars(h)
	studentID := vars["id"]        // Presuming "id" is the parameter name for student ID
	newFood := h.FormValue("food") // Assuming "food" is the parameter name for new food

	// Call repository to edit food for student
	err := r.foodServiceRepo.EditFoodForStudent(studentID, newFood)
	if err != nil {
		r.logger.Print("Database exception: ", err)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("Error updating student's food."))
		return
	}

	// Respond with success status
	rw.WriteHeader(http.StatusOK)
}


// GetAllFood returns all food items and sends them as a JSON response.
func (h *FoodServiceHandler) GetAllFood(rw http.ResponseWriter, r *http.Request) {
	foods, err := h.foodServiceRepo.GetAllFood()
	if err != nil {
		http.Error(rw, "Error retrieving food items", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(rw).Encode(foods)
	if err != nil {
		http.Error(rw, "Error encoding food items to JSON", http.StatusInternalServerError)
		return
	}
}

// getAllFood
func (r *FoodServiceHandler) GetAllFoodOfStudents(rw http.ResponseWriter, h *http.Request) {
	students, err := r.foodServiceRepo.GetAllFoodOfStudents()
	if err != nil {
		r.logger.Print("Database exception")
	}

	if students == nil {
		return
	}

	err = students.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		r.logger.Fatal("Unable to convert to json")
		return
	}
}

func (r *FoodServiceHandler) GetTherapiesFromHealthCare(rw http.ResponseWriter, h *http.Request) {
	therapies, err := r.foodServiceRepo.GetAllTherapiesFromHealthCareService()
	if err != nil {
		r.logger.Print("Database exception")
	}

	if therapies == nil {
		return
	}

	err = therapies.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		r.logger.Fatal("Unable to convert to json")
		return
	}
}

func (r *FoodServiceHandler) GetTherapies(rw http.ResponseWriter, h *http.Request) {
	therapies, err := r.foodServiceRepo.GetAllTherapiesFromFoodService()
	if err != nil {
		r.logger.Print("Error retrieving therapies from Food Service:", err)
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
		return
	}

	if therapies == nil {
		http.Error(rw, "No therapies found", http.StatusNotFound)
		return
	}

	err = therapies.ToJSON(rw)
	if err != nil {
		r.logger.Print("Error converting therapies to JSON:", err)
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *FoodServiceHandler) SaveTherapy(rw http.ResponseWriter, r *http.Request) {
	therapyData := r.Context().Value(KeyProduct{}).(*data.TherapyData)
	err := h.foodServiceRepo.SaveTherapyData(therapyData)
	if err != nil {
		h.logger.Print("Error sharing therapy data with diet service: ", err)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("Error sharing therapy data with diet service."))
		return
	}
	rw.WriteHeader(http.StatusOK)
}

/*func (h *FoodServiceHandler) EditFood(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	h.logger.Printf("Received request to edit Food with ID: %s", id)

	food, ok := r.Context().Value(KeyFood{}).(*data.Food)
	if !ok {
		h.logger.Println("Failed to retrieve Food data from context")
		http.Error(rw, "Invalid Food data", http.StatusBadRequest)
		return
	}

	h.logger.Printf("Food data retrieved: %+v", food)

	if food.FoodName == "" {
		h.logger.Println("FoodName is empty in the request")
		http.Error(rw, "FoodName cannot be empty", http.StatusBadRequest)
		return
	}

	err := h.foodServiceRepo.EditFood(id, food)
	if err != nil {
		h.logger.Printf("Error updating Food with ID %s: %v", id, err)
		http.Error(rw, "Error updating food", http.StatusInternalServerError)
		return
	}

	h.logger.Printf("Successfully updated Food with ID: %s", id)
	rw.WriteHeader(http.StatusOK)
}/*


// UpdateTherapyData ažurira podatke o terapiji u bazi podataka.



// EditFood ažurira samo FoodName za hranu u bazi podataka.
/*func (r *FoodServiceHandler) EditFood(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	id := vars["id"]

	food := h.Context().Value(KeyProduct{}).(*data.Food)

	if food.FoodName == "" {
		http.Error(rw, "FoodName cannot be empty", http.StatusBadRequest)
		return
	}

	err := r.foodServiceRepo.EditFood(id, food)
	if err != nil {
		http.Error(rw, "Error updating food", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
}*/

func (r *FoodServiceHandler) EditFood(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	id := vars["id"]

	food := h.Context().Value(KeyFood{}).(*data.Food)

	err := r.foodServiceRepo.EditFood(id, food)
	if err != nil {
		http.Error(rw, "Error updating appointment", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (h *FoodServiceHandler) DeleteFoodHandler(rw http.ResponseWriter, r *http.Request) {
	// Dobavljanje ID-a iz URL-a
	vars := mux.Vars(r)
	foodID := vars["id"]

	// Provera da li ID postoji u URL-u
	if foodID == "" {
		http.Error(rw, "Missing food ID in request", http.StatusBadRequest)
		return
	}

	// Konvertovanje string ID-a u ObjectID
	objectID, err := primitive.ObjectIDFromHex(foodID)
	if err != nil {
		http.Error(rw, "Invalid food ID format", http.StatusBadRequest)
		return
	}

	// Pozivanje funkcije koja briše hranu iz baze (preko objectID)
	err = h.foodServiceRepo.DeleteFoodEntry(objectID)
	if err != nil {
		http.Error(rw, "Error deleting food", http.StatusInternalServerError)
		return
	}

	// Ako je uspešno obrisano
	rw.WriteHeader(http.StatusNoContent) // 204 No Content
}

func (h *FoodServiceHandler) ClearTherapiesList(rw http.ResponseWriter, r *http.Request) {
	// Poziv funkcije za brisanje liste terapija iz repozitorijuma
	err := h.foodServiceRepo.ClearTherapiesCache()
	if err != nil {
		h.logger.Print("Error clearing therapies list: ", err)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("Error clearing therapies list."))
		return
	}
	// Odgovor sa statusom OK ako je brisanje uspeÄąË‡no
	rw.WriteHeader(http.StatusOK)
}

func (h *FoodServiceHandler) UpdateTherapyStatus(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	therapyID := vars["id"]
	//status := r.FormValue("status")

	var requestBody map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(rw, "Invalid request body: unable to decode JSON", http.StatusBadRequest)
	}

	status, ok := requestBody["status"].(string)
	if !ok {
		log.Println("Status nije string")
	}

	fmt.Println("Received status:", status)

	switch status {
	case data.Done, data.Undone:
		objectID, err := primitive.ObjectIDFromHex(therapyID)
		if err != nil {
			h.logger.Printf("Invalid therapy ID provided: %v", err)
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("Invalid therapy ID provided."))
			return
		}

		// Call repository to update therapy status in cache
		err = h.foodServiceRepo.UpdateTherapyStatus(objectID, data.Status(status))
		if err != nil {
			h.logger.Printf("Error updating therapy status: %v", err)
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("Error updating therapy status."))
			return
		}
		// Respond with success status
		rw.WriteHeader(http.StatusOK)
	default:
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("Invalid status provided."))
		return
	}
}

func (s *FoodServiceHandler) MiddlewareStudentDeserialization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		students := &data.Student{}
		err := students.FromJSON(h.Body)
		if err != nil {
			http.Error(rw, "Unable to decode json", http.StatusBadRequest)
			s.logger.Fatal(err)
			return
		}
		ctx := context.WithValue(h.Context(), KeyProduct{}, students)
		h = h.WithContext(ctx)
		next.ServeHTTP(rw, h)
	})
}

func (s *FoodServiceHandler) MiddlewareTherapyDeserialization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		students := &data.TherapyData{}
		err := students.FromJSON(h.Body)
		if err != nil {
			http.Error(rw, "Unable to decode json", http.StatusBadRequest)
			s.logger.Fatal(err)
			return
		}
		ctx := context.WithValue(h.Context(), KeyProduct{}, students)
		h = h.WithContext(ctx)
		next.ServeHTTP(rw, h)
	})
}

// MiddlewareFoodDeserialization je middleware funkcija koja preuzima podatke o hrani iz zahteva i stavlja ih u kontekst
func (h *FoodServiceHandler) MiddlewareFoodDeserialization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		foodData := &data.Food{}

		// Pokušaj deserializacije JSON podataka iz tela zahteva
		err := json.NewDecoder(r.Body).Decode(foodData)
		if err != nil {
			h.logger.Println("Unable to decode JSON:", err)
			http.Error(rw, "Unable to decode JSON", http.StatusBadRequest)
			return
		}

		// Postavi deserializovane podatke u kontekst
		ctx := context.WithValue(r.Context(), KeyFood{}, foodData)
		r = r.WithContext(ctx)

		// Nastavi sa sledećim handlerom
		next.ServeHTTP(rw, r)
	})
}


// MiddlewareFoodDeserialization je middleware funkcija koja preuzima podatke o hrani iz zahteva i stavlja ih u kontekst
/*func (h *FoodServiceHandler) MiddlewareFoodDeserialization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		foodData := &data.Food{}

		// Pokušaj deserializacije JSON podataka iz tela zahteva
		err := json.NewDecoder(r.Body).Decode(foodData)
		if err != nil {
			http.Error(rw, "Unable to decode JSON", http.StatusBadRequest)
			h.logger.Fatal(err)
			return
		}

		// Postavi deserializovane podatke u kontekst
		ctx := context.WithValue(r.Context(), KeyFood{}, foodData)
		r = r.WithContext(ctx)

		// Nastavi sa sledeæim handlerom
		next.ServeHTTP(rw, r)
	})
}
*/
