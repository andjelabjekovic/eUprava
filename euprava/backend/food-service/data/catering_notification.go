package data

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CateringStatus string

const (
	CATERING_NEMA CateringStatus = "NEMA"
	CATERING_IMA  CateringStatus = "IMA"
)

type NotificationState string

const (
	NOTIF_NEW   NotificationState = "NEW"
	NOTIF_SENT  NotificationState = "SENT"
	NOTIF_ERROR NotificationState = "ERROR"
)

type CateringFoodItem struct {
	Name     string `bson:"name" json:"name"`
	Type1    string `bson:"type1" json:"type1"` // PASTA|PICA|SALATA (ili šta god)
	Type2    string `bson:"type2" json:"type2"` // POSNO|MRSNO
	Quantity int    `bson:"quantity" json:"quantity"`
}

type CateringNotification struct {
	ID        primitive.ObjectID  `bson:"_id,omitempty" json:"id,omitempty"`
	RequestID string              `bson:"requestId" json:"requestId"`
	Foods     []CateringFoodItem  `bson:"foods" json:"foods"`
	Note      string              `bson:"note,omitempty" json:"note,omitempty"`

	UniversityStatus CateringStatus     `bson:"universityStatus" json:"universityStatus"` // NEMA/IMA (ono što vraćamo univerzitetu)
	State           NotificationState   `bson:"state" json:"state"`                       // NEW/SENT/ERROR
	Message         string              `bson:"message,omitempty" json:"message,omitempty"`
	CreatedAt       time.Time           `bson:"createdAt" json:"createdAt"`
	UpdatedAt       time.Time           `bson:"updatedAt" json:"updatedAt"`
	SentAt          *time.Time          `bson:"sentAt,omitempty" json:"sentAt,omitempty"`
	LastError       string              `bson:"lastError,omitempty" json:"lastError,omitempty"`
}
