package models

import "time"

type CateringFoodItem struct {
	Name     string `json:"name"`
	Type1    string `json:"type1"`
	Type2    string `json:"type2"`
	Quantity int    `json:"quantity"`
}

type CateringRequest struct {
	RequestID string             `json:"requestId"`
	Foods     []CateringFoodItem `json:"foods"`
	Note      string             `json:"note,omitempty"`
	Status    string             `json:"status"` // NEMA/IMA
	CreatedAt time.Time          `json:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt"`
}

type CreateCateringRequest struct {
	RequestID string             `json:"requestId,omitempty"` // opcionalno (ako želiš da Postman prosledi)
	Foods     []CateringFoodItem `json:"foods"`
	Note      string             `json:"note,omitempty"`
}

type UpdateCateringStatusRequest struct {
	Status string `json:"status"` // "IMA" ili "NEMA"
}
