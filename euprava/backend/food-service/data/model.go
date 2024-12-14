package data

import (
	"encoding/json"
	"io"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Firstname   string             `bson:"firstName,omitempty" json:"firstName,omitempty"`
	Lastname    string             `bson:"lastName,omitempty" json:"lastName,omitempty"`
	Gender      Gender             `bson:"gender,omitempty" json:"gender,omitempty"`
	DateOfBirth int                `bson:"date_of_birth,omitempty" json:"date_of_birth,omitempty"`
	Residence   string             `bson:"residence,omitempty" json:"residence,omitempty"`
	Email       string             `bson:"email,omitempty" json:"email,omitempty"`
	Username    string             `bson:"username,omitempty" json:"username,omitempty"`
	UserType    UserType           `bson:"userType,omitempty" json:"userType,omitempty"`
	FoodID      primitive.ObjectID `bson:"foodID,omitempty" json:"foodID,omitempty"`
}

type UserType string

const (
	MUSTERIJA = "MUSTERIJA"
	RADNIK    = "RADNIK"
)

type Gender string

const (
	Male   = "Male"
	Female = "Female"
)

type Users []*User

type Food struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID   primitive.ObjectID `bson:"userId,omitempty" json:"userId,omitempty"`
	FoodName string             `bson:"foodName,omitempty" json:"foodName,omitempty"`
}


type Foods []*Food

func (o *Food) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(o)
}

func (o *Food) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(o)
}

// ToJSON konvertuje listu hrane (Foods) u JSON format
func (o *Foods) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(o)
}

// FromJSON učitava listu hrane (Foods) iz JSON formata
func (o *Foods) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(o)
}

type AuthUser struct {
	ID            primitive.ObjectID `bson:"_id"`
	First_name    *string            `json:"first_name" validate:"required,min=2,max=100"`
	Last_name     *string            `json:"last_name" validate:"required,min=2,max=100"`
	Email         *string            `json:"email" validate:"email,required"`
	Password      *string            `json:"password" validate:"required,min=8"`
	Phone         *string            `json:"phone" validate:"required"`
	Address       *string            `json:"address" validate:"required"`
	Token         *string            `json:"token"`
	User_type     *string            `json:"user_type" validate:"required"`
	Refresh_token *string            `json:"refresh_token"`
	Created_at    time.Time          `json:"created_at"`
	Updated_at    time.Time          `json:"updated_at"`
	User_id       string             `json:"user_id"`
}

type Student struct {
	ID        primitive.ObjectID `bson:"_id" json:"id"`
	Firstname string             `bson:"firstName,omitempty" json:"firstName,omitempty"`
	Lastname  string             `bson:"lastName,omitempty" json:"lastName,omitempty"`

	//UserType  UserType           `bson:"userType" json:"userType"`
}

type Students []*Student

type TherapyData struct {
	ID        primitive.ObjectID `bson:"therapyId,omitempty" json:"therapyId,omitempty"`
	StudentID primitive.ObjectID `bson:"studentId,omitempty" json:"studentId,omitempty"`
	Diagnosis string             `bson:"diagnosis,omitempty" json:"diagnosis,omitempty"`
	Status    Status             `bson:"status,omitempty" json:"status,omitempty"`
	//Medications  []Medication       `bson:"medications,omitempty" json:"medications,omitempty"`
	//Instructions string             `bson:"instructions,omitempty" json:"instructions,omitempty"`
}
type Order struct {
	ID      primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Food    Food               `bson:"food,omitempty" json:"food,omitempty"`
	UserID  primitive.ObjectID `bson:"userId,omitempty" json:"userId,omitempty"`
	StatusO StatusO            `bson:"statusO,omitempty" json:"statusO,omitempty"`
	StatusO2 StatusO2            `bson:"statusO2,omitempty" json:"statusO2,omitempty"`

}

type StatusO string

const (
	Prihvacena   StatusO = "Prihvacena"
	Neprihvacena StatusO = "Neprihvacena"
	
)

type StatusO2 string

const (
	Otkazana   StatusO2 = "Otkazana"
	Neotkazana StatusO2 = "Neotkazana"
	
)

type Orders []*Order

func (o *Order) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(o)
}

func (o *Order) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(o)
}

func (o *Orders) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(o)
}

func (o *Orders) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(o)
}

type Status string

const (
	SentToFoodService = "sent to food service"
	Done              = "done"
	Undone            = "undone"
)

type Therapies []*TherapyData

func (o *TherapyData) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(o)
}

func (o *TherapyData) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(o)
}

func (o *Therapies) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(o)
}

func (o *Therapies) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(o)
}

func (o *Students) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(o)
}

func (o *Students) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(o)
}

func (o *Student) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(o)
}

func (o *Student) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(o)
}

func (o *Users) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(o)
}

func (o *Users) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(o)
}

func (o *User) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(o)
}

func (o *User) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(o)
}

/*
type UserType string

const (
	Guest = "Guest"
	Host  = "Host"
)*/

type UsernameChange struct {
	OldUsername string `json:"old_username"`
	NewUsername string `json:"new_username"`
}
