package data

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/sessions"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type FoodServiceRepo struct {
	cli    *mongo.Client
	logger *log.Logger
	client *http.Client
	store  *sessions.CookieStore
}

func NewFoodServiceRepo(ctx context.Context, logger *log.Logger) (*FoodServiceRepo, error) {
	dburi := fmt.Sprintf("mongodb://%s:%s/", os.Getenv("FOOD_DB_HOST"), os.Getenv("FOOD_DB_PORT"))

	client, err := mongo.NewClient(options.Client().ApplyURI(dburi))
	if err != nil {
		return nil, err
	}

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}
	store := sessions.NewCookieStore([]byte("super-secret-key"))
	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 10,
			MaxConnsPerHost:     10,
		},
	}

	// Return repository with logger and DB client
	return &FoodServiceRepo{
		logger: logger,
		cli:    client,
		client: httpClient,
		store:  store,
	}, nil
}

// Disconnect from database
func (pr *FoodServiceRepo) DisconnectMongo(ctx context.Context) error {
	err := pr.cli.Disconnect(ctx)
	if err != nil {
		return err
	}
	return nil
}

// Check database connection
func (rr *FoodServiceRepo) Ping() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check connection -> if no error, connection is established
	err := rr.cli.Ping(ctx, readpref.Primary())
	if err != nil {
		rr.logger.Println(err)
	}

	// Print available databases
	databases, err := rr.cli.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		rr.logger.Println(err)
	}
	fmt.Println(databases)
}

// GetAllFoodOfStudents
func (rr *FoodServiceRepo) GetAllFoodOfStudents() (*Students, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	studentsCollection := rr.getCollection("students")

	var students Students
	studentCursor, err := studentsCollection.Find(ctx, bson.M{})
	if err != nil {
		rr.logger.Println(err)
		return nil, err
	}
	if err = studentCursor.All(ctx, &students); err != nil {
		rr.logger.Println(err)
		return nil, err
	}
	return &students, nil
}

func (rr *FoodServiceRepo) GetTokenFromSession(r *http.Request) (string, error) {
	session, err := rr.store.Get(r, "session-name")
	if err != nil {
		return "", err
	}

	token, ok := session.Values["token"].(string)
	if !ok {
		return "", errors.New("token not found in session")
	}

	return token, nil
}

func (rr *FoodServiceRepo) GetLoggedUser(r *http.Request) (*AuthUser, error) {
	token, err := rr.GetTokenFromSession(r)
	if err != nil {
		return nil, err
	}

	meEndpoint := "http://localhost:8080/user/me"

	req, err := http.NewRequest("GET", meEndpoint, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending GET request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Unexpected status code:", resp.StatusCode)
		return nil, errors.New("unexpected status code")
	}

	var user AuthUser
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		fmt.Println("Error decoding JSON response:", err)
		return nil, err
	}

	return &user, nil
}

/*
	func (rr *FoodServiceRepo) GetAllOrdersForUser(userID primitive.ObjectID) ([]Order, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		ordersCollection := rr.getCollection("orders")

		filter := bson.M{
			"userId":   userID,
			"statusO":  "Neprihvacena",
			"statusO2": "Neotkazana",
		}

		cursor, err := ordersCollection.Find(ctx, filter)
		if err != nil {
			return nil, err
		}
		defer cursor.Close(ctx)

		var orders []Order
		if err := cursor.All(ctx, &orders); err != nil {
			rr.logger.Println(err)
			return nil, err
		}

		return orders, nil
	}
*/
func (rr *FoodServiceRepo) GetMyOrders(userID primitive.ObjectID) (Orders, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orderCollection := rr.getCollection("order") // Proveri da li je naziv kolekcije tačan

	// Filter za korisničke porudžbine koje su prihvaćene i neotkazane
	filter := bson.M{
		"userId": userID,
	}

	cursor, err := orderCollection.Find(ctx, filter)
	if err != nil {
		rr.logger.Println("Error finding user orders:", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var myOrders Orders
	if err := cursor.All(ctx, &myOrders); err != nil {
		rr.logger.Println("Error decoding user orders:", err)
		return nil, err
	}

	return myOrders, nil
}

func (rr *FoodServiceRepo) CancelOrder(orderID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orderCollection := rr.getCollection("order") // Naziv kolekcije mora biti tačan

	filter := bson.M{"_id": orderID}
	update := bson.M{"$set": bson.M{"statusO2": "Otkazana"}}

	result, err := orderCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("order not found")
	}

	return nil
}

// GetAllMyOrders vraća porudžbine korisnika sa statusO='Prihvacena' i statusO2='Neotkazana'
/*func (rr *FoodServiceRepo) GetAllMyOrders(userID primitive.ObjectID) ([]Order, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    orderCollection := rr.getCollection("order")

    // Definišemo filter za dohvat porudžbina koje pripadaju korisniku i zadovoljavaju uslove statusa
    filter := bson.M{
        "userId":  userID,
        "statusO": Prihvacena,
        "statusO2": Neotkazana,
    }

    cursor, err := orderCollection.Find(ctx, filter)
    if err != nil {
        rr.logger.Println("Error finding user's accepted and not canceled orders:", err)
        return nil, err
    }
    defer cursor.Close(ctx)

    var orders []Order
    err = cursor.All(ctx, &orders)
    if err != nil {
        rr.logger.Println("Error decoding orders:", err)
        return nil, err
    }

    rr.logger.Printf("Fetched %d orders for user %s with statusO='Prihvacena' and statusO2='Neotkazana'\n", len(orders), userID.Hex())
    return orders, nil
}*/

/*
func (rr *FoodServiceRepo) CreateFoodEntry(r *http.Request, foodData *Food) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Automatski postavljanje dummy userID (možeš ovo kasnije zameniti sa stvarnim korisničkim ID-om)
	dummyUserID := primitive.NewObjectID() // Generiše novi ObjectID
	foodData.UserID = dummyUserID          // Postavi generisani ObjectID

	// Postavi default stanje2 ako nije prosleđeno u telu zahteva
	if foodData.Stanje2 == "" {
		foodData.Stanje2 = Neprihvacena // Postavi default vrednost
	}

	// Loguj podatke pre umetanja
	fmt.Printf("Inserting food data: %+v\n", foodData)

	foodCollection := rr.getCollection("food")

	// Umetanje u MongoDB
	_, err := foodCollection.InsertOne(ctx, foodData)
	if err != nil {
		fmt.Println("Error inserting food data:", err) // Loguj grešku umetanja
		return err
	}

	// Vraćanje odgovora sa podacima
	// Pretpostavljamo da vraćaš samo foodData u odgovoru ili ga možeš modifikovati kako bi uključio userId
	return nil
}*/

func (rr *FoodServiceRepo) UpdateFoodEntry(r *http.Request, foodID primitive.ObjectID, foodData *Food) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	foodCollection := rr.getCollection("food")

	filter := bson.M{"_id": foodID}
	update := bson.M{"$set": bson.M{"foodName": foodData.FoodName}}

	_, err := foodCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}
func (rr *FoodServiceRepo) GetFoodByID(r *http.Request, id primitive.ObjectID) (*Food, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	foodCollection := rr.getCollection("food")

	filter := bson.M{"_id": id}
	var food Food
	err := foodCollection.FindOne(ctx, filter).Decode(&food)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Nema dokumenata za dati ID
		}
		return nil, err
	}

	return &food, nil
}

func (rr *FoodServiceRepo) CreateFoodEntry(r *http.Request, foodData *Food) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// Postavi jedinstveni ID za food entry
	foodData.ID = primitive.NewObjectID()

	// Nabavi kolekciju iz baze
	foodCollection := rr.getCollection("food")
	// Umetni novi unos u bazu
	_, err := foodCollection.InsertOne(ctx, foodData)
	if err != nil {
		return err
	}
	return nil
}
func (rr *FoodServiceRepo) GetAllOrders() ([]Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orderCollection := rr.getCollection("order")

	// Pripremamo prazan filter za sve dokumente
	filter := bson.M{"statusO2": Neotkazana}
	cursor, err := orderCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []Order
	err = cursor.All(ctx, &orders)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (fr *FoodServiceRepo) ApproveTherapy(therapyID primitive.ObjectID) error {
	// 1) Ažuriranje statusa u lokalnoj bazi
	if err := fr.updateTherapyStatusInDB(therapyID, Done); err != nil {
		fr.logger.Println("Error updating therapy status in DB:", err)
		return err
	}

	// 2) (Opcionalno) Slanje obaveštenja HealthCare servisu:
	if err := fr.notifyHealthCareAboutStatus(therapyID, Done); err != nil {
		fr.logger.Println("Error notifying HealthCare service:", err)
		// Možete odlučiti da li ovde vraćate grešku ili ne
		return err
	}

	return nil
}

func (fr *FoodServiceRepo) updateTherapyStatusInDB(therapyID primitive.ObjectID, newStatus Status) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	therapiesCollection := fr.getCollection("therapies")

	filter := bson.M{"_id": therapyID}
	update := bson.M{"$set": bson.M{"status": newStatus}}

	_, err := therapiesCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func (fr *FoodServiceRepo) notifyHealthCareAboutStatus(therapyID primitive.ObjectID, newStatus Status) error {
	// 1) Kreirate payload
	payload := map[string]interface{}{
		"id":     therapyID.Hex(),
		"status": newStatus,
	}
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	fr.logger.Printf("[notifyHealthCareAboutStatus] Sending payload: %s", string(jsonBytes))

	// 2) Endpoint HealthCare servisa
	healthCareHost := os.Getenv("HEALTHCARE_SERVICE_HOST")
	healthCarePort := os.Getenv("HEALTHCARE_SERVICE_PORT")
	// npr. PUT http://healthcare:8080/therapy/...
	endpoint := fmt.Sprintf("http://%s:%s/updateTherapy", healthCareHost, healthCarePort)

	// 3) HTTP request
	req, err := http.NewRequest("PUT", endpoint, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("healthcare service returned non-OK status code: %d", resp.StatusCode)
	}
	fr.logger.Printf("poslato")

	return nil
}
func (rr *FoodServiceRepo) UpdateOrderStatus(orderID primitive.ObjectID, newStatus StatusO) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orderCollection := rr.getCollection("order") // Uveri se da je naziv kolekcije tačan

	filter := bson.M{"_id": orderID}
	update := bson.M{"$set": bson.M{"statusO": newStatus}}

	result, err := orderCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("order not found")
	}

	return nil
}

// GetAcceptedOrders vraća sve porudžbine čiji je statusO='Prihvacena'
func (rr *FoodServiceRepo) GetAcceptedOrders() (Orders, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orderCollection := rr.getCollection("order") // Proveri da li je naziv kolekcije tačan

	filter := bson.M{"statusO": Prihvacena}
	cursor, err := orderCollection.Find(ctx, filter)
	if err != nil {
		rr.logger.Println("Error finding accepted orders:", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var acceptedOrders Orders
	if err := cursor.All(ctx, &acceptedOrders); err != nil {
		rr.logger.Println("Error decoding accepted orders:", err)
		return nil, err
	}

	return acceptedOrders, nil
}

func (rr *FoodServiceRepo) CreateOrderEntry(r *http.Request, orderData *Order) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Postavi jedinstveni ID za order entry
	orderData.ID = primitive.NewObjectID()

	// Postavi default status ako želiš
	orderData.StatusO = Neprihvacena
	orderData.StatusO2 = Neotkazana

	// Ako je potrebno, pronađi Food po ID-u
	if !orderData.Food.ID.IsZero() {
		foodCollection := rr.getCollection("food")
		var food Food
		err := foodCollection.FindOne(ctx, bson.M{"_id": orderData.Food.ID}).Decode(&food)
		if err != nil {
			return fmt.Errorf("Error finding food by ID: %v", err)
		}
		orderData.Food = food
	}

	orderCollection := rr.getCollection("order")
	_, err := orderCollection.InsertOne(ctx, orderData)
	if err != nil {
		return err
	}
	return nil
}

/*
// metoda kreiranje porudzbine
func (rr *FoodServiceRepo) CreateOrder(r *http.Request, orderData *Order) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Automatski postavljanje dummy userID (možeš ovo kasnije zameniti sa stvarnim korisničkim ID-om)
	dummyUserID := primitive.NewObjectID() // Generiše novi ObjectID
	orderData.UserID = dummyUserID         // Postavi generisani ObjectID
	foodId := orderData.Food.ID
	// Postavi default stanje2 ako nije prosleđeno u telu zahteva
	var food Food
	orderData.StatusO = Neprihvacena // Postavi default vrednost
	orderData.StatusO2 = StatusO2(Neotkazana)

	// Loguj podatke pre umetanja
	fmt.Printf("Inserting food data: %+v\n", orderData)

	foodCollection := rr.getCollection("food")
	orderCollection := rr.getCollection("order")
	fillter := bson.M{"_id": foodId}
	// Umetanje u MongoDB
	err := foodCollection.FindOne(ctx, fillter).Decode(&food)
	if err != nil {
		fmt.Println("Error inserting food data:", err) // Loguj grešku umetanja
		return err
	}
	orderData.Food = food
	// Umetanje u MongoDB
	_, err = orderCollection.InsertOne(ctx, orderData)
	if err != nil {
		fmt.Println("Error inserting food data:", err) // Loguj grešku umetanja
		return err
	}
	// Vraćanje odgovora sa podacima
	// Pretpostavljamo da vraćaš samo foodData u odgovoru ili ga možeš modifikovati kako bi uključio userId
	return nil
}*/

// GetListFood vraća sve unose hrane iz baze, sa dummy korisnikom (nil)
func (rr *FoodServiceRepo) GetListFood() ([]Food, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Koristi dummy UserID (nil) jer autentifikacija nije uključena
	dummyUserID := primitive.NilObjectID

	foodCollection := rr.getCollection("food")

	// Pronađi sve unose hrane bez filtriranja po korisniku
	cursor, err := foodCollection.Find(ctx, bson.M{"userId": dummyUserID})
	if err != nil {
		rr.logger.Println("Error finding food entries:", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var foodList []Food
	if err = cursor.All(ctx, &foodList); err != nil {
		rr.logger.Println("Error decoding food entries:", err)
		return nil, err
	}

	return foodList, nil
}

// GetAllFood vraća sve unose hrane iz baze podataka.
/*func (rr *FoodServiceRepo) GetAllFood() (*Foods, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
    defer cancel()

    foodCollection := rr.getCollection("food")

    var foods Foods
    cursor, err := foodCollection.Find(ctx, bson.M{})
    if err != nil {
        rr.logger.Println("Error fetching foods:", err)
        return nil, err
    }
    defer cursor.Close(ctx)

    if err = cursor.All(ctx, &foods); err != nil {
        rr.logger.Println("Error decoding foods:", err)
        return nil, err
    }

    // Dodaj log da vidiš koliko podataka je vraćeno
    rr.logger.Printf("Fetched %d foods from database", len(foods))

    return &foods, nil
}*/

// GetAllFood returns all food records from the 'food' collection.
func (rr *FoodServiceRepo) GetAllFood() (*Foods, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Access the 'food' collection.
	foodCollection := rr.getCollection("food")

	// Query to fetch all food items.
	cursor, err := foodCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Declare a variable to store the results.
	var foods Foods
	if err = cursor.All(ctx, &foods); err != nil {
		rr.logger.Println(err)
		return nil, err
	}

	// Return the food items.
	return &foods, nil
}

/*func (rr *FoodServiceRepo) EditFood(id string, food *Food) error {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	foodsCollection := rr.getCollection("foods")

	rr.logger.Printf("Starting update for Food ID: %s", id)

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		rr.logger.Println("Error converting ID to ObjectID:", err)
		return fmt.Errorf("Invalid ID format: %w", err)
	}

	rr.logger.Printf("Converted ID to ObjectID: %v", objectID)

	if food.FoodName == "" {
		rr.logger.Println("FoodName is empty, cannot proceed with update")
		return fmt.Errorf("FoodName cannot be empty")
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{
		"$set": bson.M{"foodName": food.FoodName},
	}

	rr.logger.Printf("Filter: %v, Update: %v", filter, update)

	result, err := foodsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		rr.logger.Println("Error during update:", err)
		return err
	}

	rr.logger.Printf("Documents matched: %v", result.MatchedCount)
	rr.logger.Printf("Documents updated: %v", result.ModifiedCount)

	if result.MatchedCount == 0 {
		rr.logger.Println("No documents matched the given ID.")
		return fmt.Errorf("No documents found with the provided ID")
	}

	return nil
}
*/
// edit

func (rr *FoodServiceRepo) EditFood(id string, food *Food) error {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	foodCollection := rr.getCollection("foods")

	http.DefaultClient.Timeout = 60 * time.Second

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		rr.logger.Println("Error converting ID to ObjectID:", err)
		rr.logger.Println("Invalid ID:", id)
		return err
	}

	// Ažurirajte podatke u appointmentsCollection
	filter := bson.M{"_id": objectID}
	update := bson.M{}

	if food.FoodName != "" {
		update["foodName"] = food.FoodName
	}

	updateQuery := bson.M{"$set": update}

	result, err := foodCollection.UpdateOne(ctx, filter, updateQuery)

	rr.logger.Printf("Documents matched: %v\n", result.MatchedCount)
	rr.logger.Printf("Documents updated: %v\n", result.ModifiedCount)

	if err != nil {
		rr.logger.Println(err)
		return err
	}

	return nil
}

// editFood
func (rr *FoodServiceRepo) EditFoodForStudent(studentID string, newFood string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	studentsCollection := rr.getCollection("students")

	filter := bson.M{"student_id": studentID}
	update := bson.M{"$set": bson.M{"food": newFood}}

	_, err := studentsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		rr.logger.Println(err)
		return err
	}
	rr.logger.Printf("Food updated successfully for student with ID: %s\n", studentID)
	return nil
}

// DeleteFoodEntry briše podatak o hrani iz baze podataka na osnovu ID-a.
func (rr *FoodServiceRepo) DeleteFoodEntry(foodID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	foodCollection := rr.getCollection("food")
	filter := bson.M{"_id": foodID}

	_, err := foodCollection.DeleteOne(ctx, filter)
	if err != nil {
		rr.logger.Println("Error deleting food entry:", err)
		return err
	}

	rr.logger.Printf("Successfully deleted food entry with ID: %s\n", foodID.Hex())
	return nil
}

var therapiesList Therapies

func CacheTherapies(therapies Therapies) {
	therapiesList = append(therapiesList, therapies...)
}

func GetCachedTherapies() Therapies {
	return therapiesList
}

// funkcija dobavlja sve terapije iz Food servisa.
/*func (rr *FoodServiceRepo) GetAllTherapiesFromFoodService() (Therapies, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	therapiesCollection := rr.cli.Database("MongoDatabase").Collection("therapies")

	var therapies Therapies
	cursor, err := therapiesCollection.Find(ctx, bson.M{})
	if err != nil {
		rr.logger.Println(err)
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &therapies); err != nil {
		rr.logger.Println(err)
		return nil, err
	}

	return therapies, nil
}
*/
/*func (rr *FoodServiceRepo) GetAllTherapiesFromFoodService() (Therapies, error) {
    // 1. Loguj ulazak u funkciju
    rr.logger.Println("Entering GetAllTherapiesFromFoodService")

    // 2. Kreiraj novi kontekst sa timeout-om
    rr.logger.Println("Creating context with 50 second timeout")
    ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
    defer cancel()

    // 3. Dobavi kolekciju 'therapies' iz baze 'MongoDatabase'
    rr.logger.Println("Retrieving 'therapies' collection from the 'MongoDatabase'")
    therapiesCollection := rr.cli.Database("MongoDatabase").Collection("therapies")

    // 4. Upit koji pronalazi sve dokumente
    rr.logger.Println("Finding all therapies in the database...")
    cursor, err := therapiesCollection.Find(ctx, bson.M{})
    if err != nil {
        rr.logger.Printf("Error occurred while trying to find therapies: %v\n", err)
        return nil, err
    }
    defer func() {
        rr.logger.Println("Closing cursor")
        cursor.Close(ctx)
    }()

    rr.logger.Println("Successfully retrieved cursor from the database")

    // 5. Učitaj sve rezultate iz kursora u strukturu `therapies`
    var therapies Therapies
    rr.logger.Println("Reading all documents from cursor into 'therapies'")
    if err := cursor.All(ctx, &therapies); err != nil {
        rr.logger.Printf("Error occurred while decoding cursor result: %v\n", err)
        return nil, err
    }

    rr.logger.Printf("Successfully retrieved therapies: %+v\n", therapies)

    // 6. Loguj izlazak iz funkcije
    rr.logger.Println("Leaving GetAllTherapiesFromFoodService")

    return therapies, nil
}*/
func (rr *FoodServiceRepo) GetAllTherapiesFromFoodService() (Therapies, error) {
	rr.logger.Println("[res-store] Entering GetAllTherapiesFromFoodService")

	// Kreiramo kontekst sa timeout-om
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	// Uzimamo kolekciju
	rr.logger.Println("[res-store] Retrieving 'therapies' collection from the 'MongoDatabase'")
	therapiesCollection := rr.cli.Database("MongoDatabase").Collection("therapies")

	// Pravimo upit za sve dokumente
	rr.logger.Println("[res-store] Finding all therapies in the database...")
	cursor, err := therapiesCollection.Find(ctx, bson.M{})
	if err != nil {
		rr.logger.Printf("[res-store] Error retrieving therapies from Food Service: %v", err)
		return nil, err
	}
	defer func() {
		rr.logger.Println("[res-store] Closing cursor")
		cursor.Close(ctx)
	}()

	rr.logger.Println("[res-store] Successfully retrieved cursor from the database")

	// Mapiramo rezultate iz kursora u slice struktura (Therapies)
	var therapies Therapies
	rr.logger.Println("[res-store] Reading all documents from cursor into 'therapies'")
	if err := cursor.All(ctx, &therapies); err != nil {
		rr.logger.Printf("[res-store] Error occurred while decoding cursor result: %v", err)
		return nil, err
	}

	// Izloguj sadržaj rezultata kao JSON, da vidiš šta si stvarno dobio
	b, err := json.Marshal(therapies)
	if err != nil {
		rr.logger.Printf("[res-store] Error marshaling therapies for logging: %v", err)
	} else {
		rr.logger.Printf("[res-store] Successfully retrieved therapies (JSON): %s", string(b))
	}

	rr.logger.Println("[res-store] Leaving GetAllTherapiesFromFoodService")
	return therapies, nil
}

func (rr *FoodServiceRepo) SaveTherapyData(therapyData *TherapyData) error {

	therapiesList = append(therapiesList, therapyData)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	therapiesCollection := rr.getCollection("therapies")

	// Insert therapy data into therapies collection
	_, err := therapiesCollection.InsertOne(ctx, therapyData)
	if err != nil {
		rr.logger.Println(err)
		return err
	}
	return nil
}

func (rr *FoodServiceRepo) ClearTherapiesCache() error {
	therapiesList = Therapies{}
	return nil
}

func (rr *FoodServiceRepo) UpdateTherapyStatus(therapyID primitive.ObjectID, status Status) error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	therapiesCollection := rr.getCollection("therapies")

	filter := bson.M{"therapyId": therapyID}
	update := bson.M{"$set": bson.M{"status": status}}
	result, err := therapiesCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.ModifiedCount == 0 {
		return fmt.Errorf("therapy with ID %s not found in database", therapyID.Hex())
	}

	updatedTherapy := &TherapyData{
		ID:     therapyID,
		Status: status,
	}
	if err := rr.SendTherapyDataToHealthCareService(updatedTherapy); err != nil {
		return err
	}

	return nil
}

func (rr *FoodServiceRepo) SendTherapyDataToHealthCareService(therapy *TherapyData) error {

	therapyJSON, err := json.Marshal(therapy)
	if err != nil {
		rr.logger.Println("Error serializing therapy data:", err)
		return err
	}

	healthCareHost := os.Getenv("HEALTHCARE_SERVICE_HOST")
	healthCarePort := os.Getenv("HEALTHCARE_SERVICE_PORT")
	healthCareEndpoint := fmt.Sprintf("http://%s:%s/updateTherapy", healthCareHost, healthCarePort)

	req, err := http.NewRequest("PUT", healthCareEndpoint, bytes.NewBuffer(therapyJSON))
	if err != nil {
		rr.logger.Println("Error creating request to health care service:", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := rr.client.Do(req)
	if err != nil {
		rr.logger.Println("Error sending request to health care service:", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		rr.logger.Println("Health care service returned non-OK status code:", resp.StatusCode)
		return errors.New("health care service returned non-OK status code")
	}

	return nil
}

// GetAllTherapiesFromHealthCareService funkcija dobavlja sve terapije iz HealthCare servisa.
func (rr *FoodServiceRepo) GetAllTherapiesFromHealthCareService() (Therapies, error) {
	healthCareHost := os.Getenv("HEALTHCARE_SERVICE_HOST")
	healthCarePort := os.Getenv("HEALTHCARE_SERVICE_PORT")
	healthCareEndpoint := fmt.Sprintf("http://%s:%s/therapies", healthCareHost, healthCarePort)

	req, err := http.NewRequest("GET", healthCareEndpoint, nil)
	if err != nil {
		rr.logger.Println("Error creating request to health care service:", err)
		return nil, err
	}

	resp, err := rr.client.Do(req)
	if err != nil {
		rr.logger.Println("Error sending request to health care service:", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		rr.logger.Println("Health care service returned non-OK status code:", resp.StatusCode)
		return nil, fmt.Errorf("health care service returned non-OK status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		rr.logger.Println("Error reading response from health care service:", err)
		return nil, err
	}

	var therapies Therapies
	if err := json.Unmarshal(body, &therapies); err != nil {
		rr.logger.Println("Error parsing response from health care service:", err)
		return nil, err
	}

	CacheTherapies(therapies)

	return therapies, nil
}

func (rr *FoodServiceRepo) getCollection(collectionName string) *mongo.Collection {
	return rr.cli.Database("MongoDatabase").Collection(collectionName)
}
