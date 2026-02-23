package store

import (
	"errors"
	"sync"
	"university--service/models"
)

type CateringStore struct {
	mu   sync.RWMutex
	data map[string]*models.CateringRequest
}

func NewCateringStore() *CateringStore {
	return &CateringStore{
		data: make(map[string]*models.CateringRequest),
	}
}

func (s *CateringStore) Save(req *models.CateringRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[req.RequestID] = req
}

func (s *CateringStore) Get(id string) (*models.CateringRequest, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[id]
	return v, ok
}

func (s *CateringStore) UpdateStatus(id string, status string) (*models.CateringRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.data[id]
	if !ok {
		return nil, errors.New("not found")
	}
	v.Status = status
	return v, nil
}
