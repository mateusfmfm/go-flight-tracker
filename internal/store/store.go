package store

import (
	"go-flight-tracker/internal/flight"
	"sync"
)

type AircraftStore struct {
	mu        sync.RWMutex
	aircrafts map[string]*flight.Aircraft
}

func NewAircraftStore() *AircraftStore {
	return &AircraftStore{
		aircrafts: make(map[string]*flight.Aircraft),
	}
}

func (s *AircraftStore) Update(list []*flight.Aircraft) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, a := range list {
		s.aircrafts[a.Icao24] = a
	}
}

func (s *AircraftStore) GetAll() []*flight.Aircraft {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*flight.Aircraft, 0, len(s.aircrafts))
	for _, a := range s.aircrafts {
		result = append(result, a)
	}
	return result
}

func (s *AircraftStore) GetByICAO(icao24 string) (*flight.Aircraft, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	aircraft, exists := s.aircrafts[icao24]
	return aircraft, exists
}
