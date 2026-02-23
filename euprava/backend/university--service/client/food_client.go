package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"university--service/models"
)

type FoodClient struct {
	BaseURL string
	Client  *http.Client
}

func NewFoodClient(baseURL string) *FoodClient {
	return &FoodClient{
		BaseURL: baseURL,
		Client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (fc *FoodClient) SendCateringNotification(cr *models.CateringRequest) error {
	payload := map[string]any{
		"requestId": cr.RequestID,
		"foods":     cr.Foods,
		"note":      cr.Note,
		"status":    cr.Status,
	}

	b, _ := json.Marshal(payload)
	endpoint := fc.BaseURL + "/catering/notifications"

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := fc.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("food service returned %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
