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

type HealthCareRepo struct {
	cli          *mongo.Client
	logger       *log.Logger
	client       *http.Client
	allTherapies Therapies
	store        *sessions.CookieStore
}

func NewHealthCareRepo(ctx context.Context, logger *log.Logger) (*HealthCareRepo, error) {
	dburi := fmt.Sprintf("mongodb://%s:%s/", os.Getenv("HEALTHCARE_DB_HOST"), os.Getenv("HEALTHCARE_DB_PORT"))

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
	return &HealthCareRepo{
		logger: logger,
		cli:    client,
		client: httpClient,
		store:  store,
	}, nil
}

// Disconnect from database
func (pr *HealthCareRepo) DisconnectMongo(ctx context.Context) error {
	err := pr.cli.Disconnect(ctx)
	if err != nil {
		return err
	}
	return nil
}

// Check database connection
func (rr *HealthCareRepo) Ping() {
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

// mongo
func (rr *HealthCareRepo) InsertUser(user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()
	usersCollection := rr.getCollection("users")

	// Kreiranje zdravstvenog kartona za studenta
	healthRecord := &HealthRecord{
		UserID:     user.ID,
		RecordData: "Initial health record for user " + user.Firstname + " " + user.Lastname,
	}

	insertedID, err := rr.InsertHealthRecord(healthRecord)
	if err != nil {
		return err
	}

	// Set the HealthRecordID in the user
	user.HealthRecordID = insertedID
	result, err := usersCollection.InsertOne(ctx, &user)
	if err != nil {
		rr.logger.Println(err)
		return err
	}
	rr.logger.Printf("Documents ID: %v\n", result.InsertedID)

	return nil
}

// mongo
func (rr *HealthCareRepo) InsertHealthRecord(record *HealthRecord) (primitive.ObjectID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()
	recordsCollection := rr.getCollection("health_records")

	result, err := recordsCollection.InsertOne(ctx, record)
	if err != nil {
		rr.logger.Println(err)
		return primitive.NilObjectID, err
	}
	rr.logger.Printf("Health record ID: %v\n", result.InsertedID)
	return result.InsertedID.(primitive.ObjectID), nil
}

func (rr *HealthCareRepo) GetAllUsers() (*Users, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	usersCollection := rr.getCollection("users")

	var users Users
	userCursor, err := usersCollection.Find(ctx, bson.M{})
	if err != nil {
		rr.logger.Println(err)
		return nil, err
	}
	if err = userCursor.All(ctx, &users); err != nil {
		rr.logger.Println(err)
		return nil, err
	}
	return &users, nil
}

func (rr *HealthCareRepo) GetAllHealthRecords() (*HealthRecords, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	hrCollection := rr.getCollection("health_records")

	var hr HealthRecords
	hrCursor, err := hrCollection.Find(ctx, bson.M{})
	if err != nil {
		rr.logger.Println(err)
		return nil, err
	}
	if err = hrCursor.All(ctx, &hr); err != nil {
		rr.logger.Println(err)
		return nil, err
	}
	return &hr, nil
}

// GetStudentByID vraća studenta po ID-ju.
func (rr *HealthCareRepo) GetUserByID(userID string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	usersCollection := rr.getCollection("users")

	// Konvertuj string ID u ObjectID
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %v", err)
	}

	var user User
	err = usersCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	return &user, nil
}

func (rr *HealthCareRepo) GetHealthRecordByID(hRecordId string) (*HealthRecord, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	hRecordsCollection := rr.getCollection("health_records")

	// Konvertuj string ID u ObjectID
	objID, err := primitive.ObjectIDFromHex(hRecordId)
	if err != nil {
		return nil, fmt.Errorf("invalid hRecordID ID: %v", err)
	}

	var healthRecord HealthRecord
	err = hRecordsCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&healthRecord)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("healthRecord not found")
		}
		return nil, err
	}

	return &healthRecord, nil
}

func (rr *HealthCareRepo) UpdateUser(id string, user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	usersCollection := rr.getCollection("users")

	http.DefaultClient.Timeout = 60 * time.Second

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		rr.logger.Println("Error converting ID to ObjectID:", err)
		return err
	}

	// Ažurirajte podatke u appointmentsCollection
	filter := bson.M{"_id": objectID}
	update := bson.M{}

	if user.Firstname != "" {
		update["firstName"] = user.Firstname
	}

	if user.Lastname != "" {
		update["lastName"] = user.Lastname
	}

	if user.Gender != "" {
		update["gender"] = user.Gender
	}

	if user.DateOfBirth != 0 {
		update["date_of_birth"] = user.DateOfBirth
	}

	if user.Residence != "" {
		update["residence"] = user.Residence
	}

	if user.Email != "" {
		update["email"] = user.Email
	}

	if user.Username != "" {
		update["userName"] = user.Username
	}

	if user.UserType != "" {
		update["userType"] = user.UserType
	}

	updateQuery := bson.M{"$set": update}

	result, err := usersCollection.UpdateOne(ctx, filter, updateQuery)

	rr.logger.Printf("Documents matched: %v\n", result.MatchedCount)
	rr.logger.Printf("Documents updated: %v\n", result.ModifiedCount)

	if err != nil {
		rr.logger.Println(err)
		return err
	}

	return nil
}

func (rr *HealthCareRepo) UpdateHealthRecord(id string, hRecord *HealthRecord) error {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	hRecordsCollection := rr.getCollection("health_records")

	http.DefaultClient.Timeout = 60 * time.Second

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		rr.logger.Println("Error converting ID to ObjectID:", err)
		return err
	}

	// Ažurirajte podatke u appointmentsCollection
	filter := bson.M{"_id": objectID}
	update := bson.M{}

	if hRecord.RecordData != "" {
		update["recordData"] = hRecord.RecordData
	}
	updateQuery := bson.M{"$set": update}

	result, err := hRecordsCollection.UpdateOne(ctx, filter, updateQuery)

	rr.logger.Printf("Documents matched: %v\n", result.MatchedCount)
	rr.logger.Printf("Documents updated: %v\n", result.ModifiedCount)

	if err != nil {
		rr.logger.Println(err)
		return err
	}

	return nil
}

// DeleteStudent briše studenta iz baze podataka.
func (rr *HealthCareRepo) DeleteUser(userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	usersCollection := rr.getCollection("users")

	// Konvertuj string ID u ObjectID
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %v", err)
	}

	// Kreiraj filter za pretragu studenta po ID-ju
	filter := bson.M{"_id": objID}

	_, err = usersCollection.DeleteOne(ctx, filter)
	if err != nil {
		rr.logger.Println("Error deleting user:", err)
		return err
	}

	rr.logger.Printf("Deleted user with ID: %v\n", objID)
	return nil
}

// CreateAppointment kreira novi pregled sa reserved postavljenim na false.
func (rr *HealthCareRepo) CreateAppointment(r *http.Request, appointmentData *AppointmentData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	appointmentData.ID = primitive.NewObjectID()

	examinationsCollection := rr.getCollection("examinations")

	_, err := examinationsCollection.InsertOne(ctx, appointmentData)
	if err != nil {
		return err
	}

	if appointmentData.Systematic {
		log.Printf("Slanje sistematskih podataka servisu univerziteta: %v", appointmentData)
		if err := rr.SendSystematicDataToUniversityService(appointmentData); err != nil {
			log.Printf("Greška prilikom slanja podataka servisu univerziteta: %v", err)
			return err
		}
	}

	return nil
}

// GetAllReservedAppointmentsForUser vraća sve rezervisane termine pregleda za određenog korisnika.
func (rr *HealthCareRepo) GetAllReservedAppointmentsForUser(userID primitive.ObjectID) (*Appointments, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	examinationsCollection := rr.getCollection("examinations")

	filter := bson.M{
		"reserved":   true,
		"student_id": userID,
	}

	cursor, err := examinationsCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var appointments Appointments
	if err := cursor.All(ctx, &appointments); err != nil {
		rr.logger.Println(err)
		return nil, err
	}

	return &appointments, nil
}

func (rr *HealthCareRepo) GetAllAppointmentsForUser(r *http.Request) (*Appointments, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	examinationsCollection := rr.getCollection("examinations")

	// Pronađi sve rezervisane termine pregleda za datog korisnika
	cursor, err := examinationsCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var appointments Appointments
	if err := cursor.All(ctx, &appointments); err != nil {
		rr.logger.Println(err)
		return nil, err
	}

	return &appointments, nil
}

// GetAppointmentByID dohvata pregled po ID-u.
func (rr *HealthCareRepo) GetAppointmentByID(appointmentID primitive.ObjectID) (*AppointmentData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	examinationsCollection := rr.getCollection("examinations")

	var appointment AppointmentData
	err := examinationsCollection.FindOne(ctx, bson.M{"_id": appointmentID}).Decode(&appointment)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &appointment, nil
}

func (rr *HealthCareRepo) UpdateAppointment(id string, appointment *AppointmentData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	appointmentsCollection := rr.getCollection("examinations")

	http.DefaultClient.Timeout = 60 * time.Second

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		rr.logger.Println("Error converting ID to ObjectID:", err)
		return err
	}

	// Ažurirajte podatke u appointmentsCollection
	filter := bson.M{"_id": objectID}
	update := bson.M{}

	if appointment.Date.String() != "" {
		update["date"] = appointment.Date
	}

	if appointment.DoorNumber != 0 {
		update["door_number"] = appointment.DoorNumber
	}

	if appointment.Description != "" {
		update["description"] = appointment.Description
	}

	if appointment.FacultyName != "" {
		update["faculty_name"] = appointment.FacultyName
	}

	if appointment.FieldOfStudy != "" {
		update["field_of_study"] = appointment.FieldOfStudy
	}

	if appointment.Reserved {
		update["reserved"] = appointment.Reserved
	} else if !appointment.Reserved {
		update["reserved"] = appointment.Reserved
	}

	if appointment.Systematic {
		update["systematic"] = appointment.Systematic
	} else if !appointment.Systematic {
		update["systematic"] = appointment.Systematic
	}

	updateQuery := bson.M{"$set": update}

	result, err := appointmentsCollection.UpdateOne(ctx, filter, updateQuery)

	rr.logger.Printf("Documents matched: %v\n", result.MatchedCount)
	rr.logger.Printf("Documents updated: %v\n", result.ModifiedCount)

	if err != nil {
		rr.logger.Println(err)
		return err
	}

	return nil
}

func (rr *HealthCareRepo) DeleteAppointment(appointmentID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	examinationsCollection := rr.getCollection("examinations")

	objectID, err := primitive.ObjectIDFromHex(appointmentID)
	if err != nil {
		rr.logger.Println("Error converting ID to ObjectID:", err)
		return err
	}

	var appointmentData AppointmentData
	err = examinationsCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&appointmentData)
	if err != nil {
		return err
	}

	if appointmentData.Systematic {
		appointmentData.Description = "Otkazano."

		err := rr.UpdateAppointment(appointmentID, &appointmentData)
		if err != nil {
			log.Print("Error updating appointment", http.StatusInternalServerError)
			return err
		}
		var a AppointmentData
		err = examinationsCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&appointmentData)
		if err != nil {
			return err
		}
		fmt.Println(a)

		log.Printf("Slanje sistematskih podataka servisu univerziteta: %v", appointmentData)
		if err := rr.SendSystematicDataToUniversityService(&appointmentData); err != nil {
			log.Printf("Greška prilikom slanja podataka servisu univerziteta: %v", err)
			return err
		}
	}

	_, err = examinationsCollection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}

	return nil
}

// ScheduleAppointment zakazuje pregled za određenog studenta.
func (rr *HealthCareRepo) ScheduleAppointment(r *http.Request, appointmentID, userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := rr.CreateHealthRecord(userID)
	if err != nil {
		return err
	}

	examinationsCollection := rr.getCollection("examinations")

	filter := bson.M{
		"_id": appointmentID,
	}

	update := bson.M{
		"$set": bson.M{
			"reserved":   true,
			"student_id": userID,
		},
	}

	result, err := examinationsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		rr.logger.Println("Greška pri zakazivanju pregleda:", err)
		return err
	}

	// Ako nije nađen nijedan pregled koji odgovara filteru
	if result.MatchedCount == 0 {
		return errors.New("pregled nije pronađen")
	}

	rr.logger.Printf("Uspešno zakazan pregled sa ID: %v\n", appointmentID)
	return nil
}

// Kreira zdravstveni karton za korisnika ako ne postoji.
func (rr *HealthCareRepo) CreateHealthRecord(userID primitive.ObjectID) error {
	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	exists, err := rr.CheckHealthRecordExists(userID)
	if err != nil {
		rr.logger.Println("Greška pri proveri zdravstvenog kartona:", err)
		return err
	}

	if exists {
		return nil
	}

	healthRecord := &HealthRecord{
		UserID:     userID,
		RecordData: "Initial health record for user",
	}

	insertedID, err := rr.InsertHealthRecord(healthRecord)
	if err != nil {
		return err
	}

	rr.logger.Printf("Uspešno kreiran zdravstveni karton za korisnika sa ID: %v\n, %v\n", userID, insertedID)
	return nil
}

// CancelAppointment ažurira pregled za određenog studenta tako da je reserved postavljen na false.
func (rr *HealthCareRepo) CancelAppointment(appointmentID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	examinationsCollection := rr.getCollection("examinations")

	// Kreiraj filter za pretragu pregleda po ID-ju
	filter := bson.M{"_id": appointmentID}

	// Kreiraj update za postavljanje reserved na false
	update := bson.M{
		"$set": bson.M{
			"reserved":  false,
			"studentId": nil, // ili prazna vrednost u zavisnosti od tvojih zahteva
		},
	}

	// Pokušaj ažuriranja dokumenta
	result, err := examinationsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		rr.logger.Println("Error cancelling appointment:", err)
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("appointment not found")
	}

	rr.logger.Printf("Cancelled appointment with ID: %v\n", appointmentID)
	return nil
}

// GetAllReservedAppointments vraća sve rezervisane termine pregleda.
func (rr *HealthCareRepo) GetAllReservedAppointments() (*Appointments, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	examinationsCollection := rr.getCollection("examinations")
	cursor, err := examinationsCollection.Find(ctx, bson.M{"reserved": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var appointments Appointments
	if err = cursor.All(ctx, &appointments); err != nil {
		rr.logger.Println(err)
		return nil, err
	}
	return &appointments, nil
}

// GetAllNotReservedAppointments vraća sve nerezevisane termine pregleda.
func (rr *HealthCareRepo) GetAllNotReservedAppointments() (*Appointments, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	examinationsCollection := rr.getCollection("examinations")

	filter := bson.M{
		"reserved": false,
		"date":     bson.M{"$gte": time.Now()},
	}

	cursor, err := examinationsCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var appointments Appointments
	if err = cursor.All(ctx, &appointments); err != nil {
		rr.logger.Println(err)
		return nil, err
	}
	return &appointments, nil
}

// GetAllAppointments vraća sve termine pregleda.
func (rr *HealthCareRepo) GetAllAppointments() (*Appointments, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	examinationsCollection := rr.getCollection("examinations")
	// Pronađi sve termine pregleda
	cursor, err := examinationsCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var appointments Appointments
	userCursor, err := examinationsCollection.Find(ctx, bson.M{})
	if err != nil {
		rr.logger.Println(err)
		return nil, err
	}
	if err = userCursor.All(ctx, &appointments); err != nil {
		rr.logger.Println(err)
		return nil, err
	}
	return &appointments, nil
}

// SaveTherapyData funkcija čuva podatke o terapiji u bazi podataka
func (rr *HealthCareRepo) SaveTherapyData(therapyData *TherapyData) (primitive.ObjectID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	therapiesCollection := rr.getCollection("therapies")

	// Insert therapy data into therapies collection
	result, err := therapiesCollection.InsertOne(ctx, therapyData)
	if err != nil {
		rr.logger.Println(err)
		return primitive.NilObjectID, err
	}

	// Vraća ID umetnutog dokumenta
	insertedID := result.InsertedID.(primitive.ObjectID)
	// **Bitno**: upišemo ga nazad u struct da bismo ga kasnije prosledili Food Service-u
	therapyData.ID = insertedID
	return insertedID, nil
}

func (rr *HealthCareRepo) UpdateTherapyData(id string, therapy *TherapyData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	therapiesCollection := rr.getCollection("therapies")

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

	if therapy.Diagnosis != "" {
		update["diagnosis"] = therapy.Diagnosis
	}

	if therapy.Status != "" {
		update["status"] = therapy.Status
	}

	updateQuery := bson.M{"$set": update}

	result, err := therapiesCollection.UpdateOne(ctx, filter, updateQuery)

	rr.logger.Printf("Documents matched: %v\n", result.MatchedCount)
	rr.logger.Printf("Documents updated: %v\n", result.ModifiedCount)

	if err != nil {
		rr.logger.Println(err)
		return err
	}

	return nil
}

// DeleteTherapyData briše podatke o terapiji iz baze podataka.
func (rr *HealthCareRepo) DeleteTherapyData(therapyID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	therapiesCollection := rr.getCollection("therapies")
	filter := bson.M{"_id": therapyID}

	_, err := therapiesCollection.DeleteOne(ctx, filter)
	if err != nil {
		rr.logger.Println("Error deleting therapy data:", err)
		return err
	}

	return nil
}

// GetTherapyDataByID vraća podatke o terapiji na osnovu ID-a.
func (rr *HealthCareRepo) GetTherapyDataByID(therapyID primitive.ObjectID) (*TherapyData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	therapiesCollection := rr.getCollection("therapies")
	filter := bson.M{"_id": therapyID}

	var therapyData TherapyData
	err := therapiesCollection.FindOne(ctx, filter).Decode(&therapyData)
	if err != nil {
		rr.logger.Println("Error getting therapy data by ID:", err)
		return nil, err
	}

	return &therapyData, nil
}

// GetAllTherapies dohvata sve terapije iz baze podataka.
func (rr *HealthCareRepo) GetAllTherapies() (Therapies, error) {
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

// ShareTherapyDataWithDietService funkcija deli informacije o terapijama sa službom ishrane.
func (rr *HealthCareRepo) SaveAndShareTherapyDataWithDietService(therapyData *TherapyData) error {
	// 1. Zapise podatke o terapiji
	id, err := rr.SaveTherapyData(therapyData)
	if err != nil {
		return err
	}

	// Dodaj terapiju u listu svih terapija
	rr.allTherapies = append(rr.allTherapies, therapyData)

	// 2. Pošalje podatke o terapiji službi ishrane
	if err := rr.SendTherapyDataToDietService(id, therapyData); err != nil {
		return err
	}

	return nil
}

// SendTherapyDataToDietService funkcija šalje podatke o terapiji službi ishrane
func (rr *HealthCareRepo) SendTherapyDataToDietService(id primitive.ObjectID, therapyData *TherapyData) error {
	// Serializacija terapije u JSON
	therapyJSON, err := json.Marshal(therapyData)
	if err != nil {
		rr.logger.Println("Error serializing therapy data:", err)
		return err
	}

	// Konfiguracija endpointa servisa ishrane
	foodServiceHost := os.Getenv("FOOD_SERVICE_HOST")
	foodServicePort := os.Getenv("FOOD_SERVICE_PORT")
	foodServiceEndpoint := fmt.Sprintf("http://%s:%s/therapy", foodServiceHost, foodServicePort)

	// Kreiranje HTTP zahteva
	req, err := http.NewRequest("POST", foodServiceEndpoint, bytes.NewBuffer(therapyJSON))
	if err != nil {
		rr.logger.Println("Error creating request to food service:", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// Slanje zahteva servisu ishrane
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		rr.logger.Println("Error sending request to food service:", err)
		return err
	}
	defer resp.Body.Close()

	// Provera status koda odgovora
	if resp.StatusCode != http.StatusOK {
		rr.logger.Println("Food service returned non-OK status code:", resp.StatusCode)
		return errors.New("food service returned non-OK status code")
	}

	return nil
}

// Proverava da li postoji HealthRecordID
func (rr *HealthCareRepo) CheckHealthRecordExists(userId primitive.ObjectID) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()
	recordsCollection := rr.getCollection("health_records")

	filter := bson.M{"userId": userId}
	count, err := recordsCollection.CountDocuments(ctx, filter)
	if err != nil {
		rr.logger.Println("Error checking health record existence:", err)
		return false, err
	}

	return count > 0, nil
}

func (rr *HealthCareRepo) UpdateTherapyFromFoodService(updatedTherapy *TherapyData) error {

	var existingTherapy *TherapyData
	for _, therapy := range rr.allTherapies {
		if therapy.ID.Hex() == updatedTherapy.ID.Hex() {
			existingTherapy = therapy
			break
		}

	}

	if existingTherapy == nil {
		return fmt.Errorf("therapy with ID %s not found", updatedTherapy.ID)
	}

	existingTherapy.Status = updatedTherapy.Status

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	therapiesCollection := rr.getCollection("therapies")

	filter := bson.M{"_id": updatedTherapy.ID}
	update := bson.M{"$set": bson.M{
		"status": existingTherapy.Status,
	}}

	_, err := therapiesCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		rr.logger.Println("Error updating therapy in database:", err)
		return err
	}

	rr.logger.Printf("promenjeno")
	return nil
}

// Funkcija koja dobavlja terapije koje su završene sa servera za ishranu
func (rr *HealthCareRepo) GetDoneTherapiesFromFoodService() (Therapies, error) {
	// Konstruisanje URL endpointa za dobavljanje terapija koje su završene
	foodServiceHost := os.Getenv("FOOD_SERVICE_HOST")
	foodServicePort := os.Getenv("FOOD_SERVICE_PORT")
	foodServiceEndpoint := fmt.Sprintf("http://%s:%s/therapies/done", foodServiceHost, foodServicePort)

	// Kreiranje HTTP GET zahteva na odgovarajući endpoint
	req, err := http.NewRequest("GET", foodServiceEndpoint, nil)
	if err != nil {
		rr.logger.Println("Error creating request to food service:", err)
		return nil, err
	}

	// Slanje zahteva serveru za ishranu
	resp, err := rr.client.Do(req)
	if err != nil {
		rr.logger.Println("Error sending request to food service:", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Provera status koda odgovora
	if resp.StatusCode != http.StatusOK {
		rr.logger.Println("Food service returned non-OK status code:", resp.StatusCode)
		return nil, errors.New("food service returned non-OK status code")
	}

	// Čitanje odgovora
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		rr.logger.Println("Error reading response from food service:", err)
		return nil, err
	}

	// Parsiranje odgovora u listu terapija
	var therapies Therapies
	if err := json.Unmarshal(body, &therapies); err != nil {
		rr.logger.Println("Error parsing response from food service:", err)
		return nil, err
	}

	return therapies, nil
}

func (rr *HealthCareRepo) SendSystematicDataToUniversityService(appointment *AppointmentData) error {

	appointmentJSON, err := json.Marshal(appointment)
	if err != nil {
		rr.logger.Println("Error serializing appointment data:", err)
		return err
	}

	universityServiceHost := os.Getenv("UNIVERSITY_SERVICE_HOST")
	universityServicePort := os.Getenv("UNIVERSITY_SERVICE_PORT")
	universityServiceEndpoint := fmt.Sprintf("http://%s:%s/notificationsByHealthcare", universityServiceHost, universityServicePort)

	println(universityServiceEndpoint)

	req, err := http.NewRequest("POST", universityServiceEndpoint, bytes.NewBuffer(appointmentJSON))
	if err != nil {
		rr.logger.Println("Error creating request to university service:", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// Šaljemo zahtev servisu ishrane
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		rr.logger.Println("Error sending request to university service:", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		rr.logger.Println("University service returned non-OK status code:", resp.StatusCode)
		return errors.New("University service returned non-OK status code")
	}

	return nil
}

func (rr *HealthCareRepo) getCollection(collectionName string) *mongo.Collection {
	return rr.cli.Database("MongoDatabase").Collection(collectionName)
}
